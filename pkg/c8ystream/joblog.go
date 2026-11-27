package c8ystream

import (
	"bytes"
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

type workerIDKey struct{}

// withWorkerID returns ctx carrying the number (1 based) of the worker which runs the job
func withWorkerID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, workerIDKey{}, id)
}

func workerID(ctx context.Context) int {
	if id, ok := ctx.Value(workerIDKey{}).(int); ok {
		return id
	}
	return 1
}

// logJobs wraps the producer so that --verbose logs the jobs as v2.59.0 did:
//
//	Max jobs: 0
//	worker 1: started job 1
//	...
//	worker 1: finished job 1 in 6ms
//
// --debug additionally logs the start of the workers and the outcome of each job.
func (r *Runner) logJobs(produce Producer) Producer {
	log, err := r.Factory.Logger()
	if err != nil || log == nil {
		return produce
	}
	workers := max(r.Config.GetWorkers(), 1)
	for id := 1; id <= workers; id++ {
		log.Debugf("starting worker: %d", id)
	}
	log.Infof("Max jobs: %d", r.Config.GetMaxJobs())

	var jobCount atomic.Int64
	return func() (Job, error) {
		job, err := produce()
		if err != nil || job == nil {
			return job, err
		}
		jobID := jobCount.Add(1)
		log.Debugf("adding job: %d", jobID)
		return func(ctx context.Context) output.Seq {
			return func(yield func(jsondoc.JSONDoc, error) bool) {
				worker := workerID(ctx)
				log.Infof("worker %d: started job %d", worker, jobID)
				start := time.Now()
				finished := false
				finish := func(jobErr error) {
					if finished {
						return
					}
					finished = true
					log.Infof("worker %d: finished job %d in %dms", worker, jobID, time.Since(start).Milliseconds())
					if jobErr != nil {
						log.Infof("job error. %s", r.toServerError(jobErr))
					} else {
						log.Debug("job successful")
					}
				}
				defer finish(nil)
				for doc, err := range job(ctx) {
					if err != nil {
						// the error is reported while it is yielded: log the end of the job before
						finish(err)
					}
					if !yield(doc, err) {
						return
					}
				}
			}
		}, nil
	}
}

// htmlEntities are the escapes of "<", ">" and "&" which the output is cleaned of
// (the backslash is not written as part of the literal: it would be read as an escape)
var htmlEntities = []struct{ escaped, literal string }{
	{"\\" + "u003c", "<"},
	{"\\" + "u003e", ">"},
	{"\\" + "u0026", "&"},
}

// unescapeHTML replaces the escaped "<", ">" and "&" of the output text with the
// characters, as v2.59.0 did for all output.
//
// v2.59.0 replaced the six characters wherever they were, also after an escaped
// backslash, which turns valid JSON into invalid JSON (e.g. the JSON string "a\\u003c").
// The renderers parse the JSON (e.g. csv), so in this case only the real escapes are
// replaced and the document stays valid.
func unescapeHTML(b []byte) []byte {
	if !bytes.Contains(b, []byte("\\"+"u00")) {
		return b
	}
	out := b
	for _, e := range htmlEntities {
		out = bytes.ReplaceAll(out, []byte(e.escaped), []byte(e.literal))
	}
	if len(out) != len(b) && !json.Valid(out) && json.Valid(b) {
		return unescapeHTMLEscapes(b)
	}
	return out
}

// unescapeHTMLEscapes replaces the escapes of "<", ">" and "&" of a JSON text, but not
// the text which follows an escaped backslash
func unescapeHTMLEscapes(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] != '\\' || i+1 >= len(b) {
			out = append(out, b[i])
			continue
		}
		replaced := false
		for _, e := range htmlEntities {
			if bytes.HasPrefix(b[i:], []byte(e.escaped)) {
				out = append(out, e.literal...)
				i += len(e.escaped) - 1
				replaced = true
				break
			}
		}
		if !replaced {
			// an escape sequence (e.g. an escaped backslash) is kept as it is
			out = append(out, b[i], b[i+1])
			i++
		}
	}
	return out
}

// unescapeHTMLStage is the last stage before the output is rendered
func unescapeHTMLStage() output.Stage {
	return output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		raw := doc.Raw()
		if out := unescapeHTML(raw); len(out) != len(raw) {
			return jsondoc.New(out), nil
		}
		return doc, nil
	})
}
