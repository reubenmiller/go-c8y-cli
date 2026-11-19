// Package c8ystream bridges the CLI input machinery (pipeline input, flag
// iterators, templates) to go-c8y v2 service calls and the v2 output
// pipeline.
//
// The model: a Producer drains the command's input machinery one item at a
// time and returns a Job per item — a closure over the already-resolved
// per-item values (query, id, body) that performs the service call and
// yields the resulting documents. Run executes the jobs (optionally on a
// worker pool) and flattens every produced document into a single
// output.Seq, which feeds output.Render together with the standard
// client-side stages (filter, output template, select).
//
// This replaces the requestiterator.RequestIterator + worker.Worker pairing
// for v2-based commands: instead of materialising a raw HTTP request per
// input item, the producer materialises a typed service call per input item.
package c8ystream

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// Job performs one unit of work (typically one service call resolved from a
// single input item) and yields the documents it produced. A failing job
// yields its error into the stream; Run does not inspect job internals.
type Job func(ctx context.Context) output.Seq

// Producer returns the next Job, resolved from the command's input machinery
// (query/path templates, body builders). io.EOF ends iteration. An error
// wrapping cmderrors.ErrNoMatchesFound is surfaced as an item-level error and
// iteration continues (matching requestiterator.RequestIterator); any other
// error is surfaced and ends iteration.
type Producer func() (Job, error)

// Limit caps a producer at n jobs. n <= 0 means no limit.
func Limit(p Producer, n int64) Producer {
	if n <= 0 {
		return p
	}
	count := int64(0)
	return func() (Job, error) {
		if count >= n {
			return nil, io.EOF
		}
		count++
		return p()
	}
}

// FromResult adapts a single typed service result into a document stream.
// Results without a body (e.g. 204 No Content) yield nothing.
func FromResult[T jsondoc.Unwrapper](res op.Result[T]) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if res.Err != nil {
			yield(jsondoc.Empty(), res.Err)
			return
		}
		doc := res.Data.GetJSONDoc()
		if len(doc.Raw()) == 0 {
			return
		}
		yield(doc, nil)
	}
}

// FromResponse adapts a typed service result by rendering its raw response body
// exactly as received from the server (op.Result.Response), without plucking the
// collection items into Data. Use this for endpoints backed by a collection call
// whose CLI output must be kept as-is — the spec's collectionProperty "-" — such
// as the application get-by-version/tag endpoints, which the server may answer
// with either the whole collection envelope or a single object. Falls back to
// Data when the raw body was not retained (single-object Execute results), and
// yields nothing under dry run (the request is rendered by the dry-run handler).
func FromResponse[T jsondoc.Unwrapper](res op.Result[T]) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if res.Err != nil {
			yield(jsondoc.Empty(), res.Err)
			return
		}
		if len(res.Response) > 0 {
			yield(jsondoc.New(res.Response), nil)
			return
		}
		doc := res.Data.GetJSONDoc()
		if len(doc.Raw()) == 0 {
			return
		}
		yield(doc, nil)
	}
}

// FromStatus adapts a result with no renderable body (e.g. Delete returning
// core.NoContent): it surfaces an error, and yields nothing on success.
func FromStatus[T any](res op.Result[T]) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if res.Err != nil {
			yield(jsondoc.Empty(), res.Err)
		}
	}
}

// FromJSONValue adapts a typed service result whose body is a plain Go value
// rather than a jsondoc model (e.g. the map[string]string returned by the
// by-category tenant-option endpoints) into a document stream by JSON-encoding
// res.Data. Pass dry=apiv2.IsDryRun(ctx): under --dry-run the request was
// already rendered by the dry-run handler, so the zero-value result is
// suppressed instead of emitting a spurious placeholder.
func FromJSONValue[T any](res op.Result[T], dry bool) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if res.Err != nil {
			yield(jsondoc.Empty(), res.Err)
			return
		}
		if dry {
			return
		}
		raw, err := json.Marshal(res.Data)
		if err != nil {
			yield(jsondoc.Empty(), err)
			return
		}
		yield(jsondoc.New(raw), nil)
	}
}

// FromDocs yields the given locally-produced documents. Use for non-API
// commands so their output flows through the same pipeline as API responses.
func FromDocs(docs ...jsondoc.JSONDoc) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		for _, doc := range docs {
			if !yield(doc, nil) {
				return
			}
		}
	}
}

// Stream drains the producer, executes each job and flattens all yielded
// documents into one stream.
//
// workers <= 1 keeps the pipeline fully pull-based and lazy: stopping the
// consumer (e.g. output.Head) stops pagination and input consumption alike,
// and no goroutines are involved.
//
// workers > 1 executes jobs on a bounded worker pool. The producer is always
// drained serially (the CLI input machinery is not safe for concurrent use);
// only job execution is concurrent. Documents are merged in arrival order.
// Stopping the consumer cancels the context governing in-flight jobs, though
// requests already on the wire run to completion.
func Stream(ctx context.Context, workers int, produce Producer) output.Seq {
	if workers <= 1 {
		return runSerial(ctx, produce)
	}
	return runConcurrent(ctx, workers, produce)
}

func runSerial(ctx context.Context, produce Producer) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		for {
			job, err := produce()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				if !yield(jsondoc.Empty(), err) {
					return
				}
				if errors.Is(err, cmderrors.ErrNoMatchesFound) {
					continue
				}
				return
			}
			for doc, err := range job(ctx) {
				if !yield(doc, err) {
					return
				}
			}
		}
	}
}

func runConcurrent(ctx context.Context, workers int, produce Producer) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		ctx, cancel := context.WithCancel(ctx)
		// Consumer stopped (Head, render error, ...): stop the producer and
		// in-flight jobs. Requests already sent still complete server-side.
		defer cancel()

		type item struct {
			doc jsondoc.JSONDoc
			err error
		}
		jobs := make(chan Job)
		results := make(chan item, workers)

		go func() {
			defer close(jobs)
			for {
				job, err := produce()
				if errors.Is(err, io.EOF) {
					return
				}
				if err != nil {
					// Route producer errors through the job channel so they
					// reach the consumer in stream order.
					stop := !errors.Is(err, cmderrors.ErrNoMatchesFound)
					errJob := Job(func(context.Context) output.Seq {
						return func(yield func(jsondoc.JSONDoc, error) bool) {
							yield(jsondoc.Empty(), err)
						}
					})
					select {
					case jobs <- errJob:
					case <-ctx.Done():
						return
					}
					if stop {
						return
					}
					continue
				}
				select {
				case jobs <- job:
				case <-ctx.Done():
					return
				}
			}
		}()

		var wg sync.WaitGroup
		for range workers {
			wg.Go(func() {
				for job := range jobs {
					for doc, err := range job(ctx) {
						select {
						case results <- item{doc, err}:
						case <-ctx.Done():
							return
						}
					}
				}
			})
		}
		go func() {
			wg.Wait()
			close(results)
		}()

		for it := range results {
			if !yield(it.doc, it.err) {
				return
			}
		}
	}
}
