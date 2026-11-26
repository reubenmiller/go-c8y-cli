package c8ystream

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// BatchItem resolves one input item to the value it adds to a batch (for
// example the item's body).
type BatchItem func(in *Resolver) (json.RawMessage, error)

// BatchCall returns the call which sends one batch of item values.
type BatchCall func(items []json.RawMessage) Call

// RunBatch is Run for a command which sends its input items in batches (for
// example a bulk create): item resolves each input item, and send returns the
// call for every size items (the last batch can be smaller). An item which
// cannot be resolved is reported as an error and left out of its batch.
// Without piped input the command sends one batch of one item. --maxJobs
// limits the number of batches, and the input of --outputTemplate is the
// batch's last item.
func (r *Runner) RunBatch(size int, item BatchItem, send BatchCall) error {
	return r.run(func() Producer { return r.batchProducer(size, item, send) })
}

func (r *Runner) batchProducer(size int, item BatchItem, send BatchCall) Producer {
	if size < 1 {
		size = 1
	}
	done := false
	produce := Producer(func() (Job, error) {
		var values []json.RawMessage
		var errs []error
		for !done && len(values)+len(errs) < size {
			in, err := r.nextItem()
			if errors.Is(err, io.EOF) {
				done = true
				break
			}
			if err != nil {
				return nil, err
			}
			if r.driver == nil {
				// Without a driver the command runs once
				done = true
			}
			value, err := item(in)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			values = append(values, value)
		}
		if len(values) == 0 && len(errs) == 0 {
			return nil, io.EOF
		}

		var call Call
		if len(values) > 0 {
			call = send(values)
		}
		job := Job(func(ctx context.Context) output.Seq {
			return func(yield func(jsondoc.JSONDoc, error) bool) {
				for _, err := range errs {
					if !yield(jsondoc.Empty(), err) {
						return
					}
				}
				if call == nil {
					return
				}
				// One approval scope per batch: a batch is one request
				for doc, err := range r.adaptItem(call(apiv2.WithConflictAsError(apiv2.WithNotFoundAsError(apiv2.WithApprovalScope(ctx)))), "") {
					if !yield(doc, err) {
						return
					}
				}
			}
		})
		if r.submission != nil {
			job = r.submission.withDelays(job)
		}
		if r.tmpl != nil {
			job = r.wrapTemplateJob(r.tmpl, r.tmplFlags, job)
		}
		return job, nil
	})
	return Limit(produce, r.Config.GetMaxJobs())
}
