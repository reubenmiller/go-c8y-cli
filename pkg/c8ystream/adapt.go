package c8ystream

import (
	"errors"
	"fmt"
	"iter"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// pageSource is a paginating result that can also yield whole pages (the
// response envelopes), used for --raw.
type pageSource interface {
	PageDocuments() iter.Seq2[jsondoc.JSONDoc, error]
}

// adapt turns the value a Call returned into the document stream every command
// renders the same way:
//
//   - a pagination iterator yields its items, or its pages with --raw;
//   - an op.Result yields its document (nothing for 204, dry run, deferred);
//   - an output.Seq (local documents) passes through;
//   - an error is yielded as an item error.
//
// A request the user declined at the confirmation prompt is skipped silently
// (no output, no error), matching the v1 "no" answer.
func (r *Runner) adapt(v any) output.Seq {
	var seq output.Seq
	switch x := v.(type) {
	case nil:
		return emptySeq
	case output.Seq:
		seq = x
	case func(yield func(jsondoc.JSONDoc, error) bool):
		seq = x
	case error:
		seq = errSeq(x)
	case pageSource:
		if r.Config.RawOutput() {
			seq = output.Seq(x.PageDocuments())
		} else if ds, ok := v.(op.DocumentSource); ok {
			seq = output.Seq(ds.Documents())
		} else {
			seq = output.Seq(x.PageDocuments())
		}
	case op.DocumentSource:
		seq = output.Seq(x.Documents())
	default:
		return errSeq(fmt.Errorf("c8ystream: unsupported call result %T", v))
	}
	return skipDeclined(seq)
}

// approvalError marks an error returned by the confirmation prompt (e.g. "No"
// or "No to all"), so it can be reported without the HTTP error wrapping the
// SDK transport adds.
type approvalError struct{ err error }

func (e approvalError) Error() string { return e.err.Error() }
func (e approvalError) Unwrap() error { return e.err }

// skipDeclined drops the error of a request declined without an error, and
// unwraps the prompt's own error from the transport error of a request that
// was declined with one.
func skipDeclined(seq output.Seq) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		for doc, err := range seq {
			if err != nil {
				var ae approvalError
				if errors.As(err, &ae) {
					err = ae.err
				} else if errors.Is(err, apiv2.ErrRequestNotApproved) {
					continue
				}
			}
			if !yield(doc, err) {
				return
			}
		}
	}
}
