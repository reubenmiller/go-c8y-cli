package c8ystream

import (
	"bytes"
	"context"
	"net/http"
	"sync/atomic"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// wrapRawFileJob writes the response of a job to the --outputFileRaw path, as
// received from the server and before any client-side processing (the way the
// v1 request handler wrote resp.Body()): the whole body of a list, with its
// collection property and statistics, and the last page of several. The
// responses are the ones which the output template capture records.
//
// A document which has no response of its own (a result which was not read
// from a JSON response) is written as it is. A call which writes the file
// itself (a binary download, or a response which is not JSON, see
// markRawFileHandled) is left alone.
//
// The path may contain the placeholders of rawFileFields ({id}, {name},
// {type}, {owner}, {filename} and {basename}), so a stream of items writes one
// file per item (e.g. "out/{id}.json"). Writes truncate (not append), matching
// v1: when several responses resolve to the same path the last one wins.
func (r *Runner) wrapRawFileJob(pathTemplate string, job Job) Job {
	return func(ctx context.Context) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			capture, ok := ctx.Value(templateCaptureKey{}).(*templateCapture)
			if !ok {
				capture = &templateCapture{}
				ctx = context.WithValue(ctx, templateCaptureKey{}, capture)
			}
			state := &rawFileState{}
			ctx = context.WithValue(ctx, rawFileKey{}, state)
			failed, written := false, false
			job(ctx)(func(doc jsondoc.JSONDoc, err error) bool {
				if err != nil {
					failed = true
					return yield(doc, err)
				}
				if pages, _, _ := capture.get(); len(pages) == 0 && !state.handled.Load() {
					raw := doc.Raw()
					if werr := writeRawFile(pathTemplate, raw, rawFileFields(nil, raw)); werr != nil {
						failed = true
						return yield(jsondoc.Empty(), werr)
					}
					written = true
				}
				return yield(doc, nil)
			})
			if failed || written || state.handled.Load() {
				return
			}
			pages, _, _ := capture.get()
			last, ok := capture.last()
			if len(pages) > 0 {
				last, ok = pages[len(pages)-1], true
			} else if ok && len(last.body) > 0 {
				// the response was written by the call or as its documents
				ok = false
			}
			// the last page of a list, or the empty response of a request (e.g.
			// a delete), but not one of a dry run
			if ok && !last.dry {
				if err := writeRawFile(pathTemplate, last.body, last.fields); err != nil {
					yield(jsondoc.Empty(), err)
				}
			}
		}
	}
}

// rawFileKey is the context key of the rawFileState of a job.
type rawFileKey struct{}

// rawFileState tells whether the call of a job wrote --outputFileRaw itself.
type rawFileState struct {
	handled atomic.Bool
}

// markRawFileHandled records that the call of the job of ctx wrote
// --outputFileRaw itself, so wrapRawFileJob does not write it again.
func markRawFileHandled(ctx context.Context) {
	if state, ok := ctx.Value(rawFileKey{}).(*rawFileState); ok {
		state.handled.Store(true)
	}
}

func writeRawFile(pathTemplate string, raw []byte, fields map[string]string) error {
	_, err := fileutilities.WriteToFile(bytes.NewReader(raw), pathTemplate, false, true, fields)
	return err
}

// installOutputFileCapture records that a response with a body was received,
// which makes --outputFile write an empty file for a list with no items (as
// v2.59.0 did), but not for a failed request, a dry run or a response with no
// body.
func (r *Runner) installOutputFileCapture(client *apiv2.Client) {
	if client == nil || r.Config.GetOutputFile() == "" {
		return
	}
	client.Use(observeResponses(func(o observedResponse) {
		if o.Response.StatusCode < http.StatusBadRequest && len(o.Body) > 0 {
			r.fileReceived.Store(true)
		}
	}))
}
