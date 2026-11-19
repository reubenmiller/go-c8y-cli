package c8ystream

import (
	"errors"
	"fmt"
	"io"
	"iter"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"

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
//   - a binary download is written to --outputFileRaw, or to stdout;
//   - an output.Seq (local documents) passes through;
//   - an error is yielded as an item error.
//
// A request the user declined at the confirmation prompt is skipped silently
// (no output, no error), matching the v1 "no" answer.
func (r *Runner) adapt(v any) output.Seq {
	return r.adaptItem(v, "")
}

// adaptItem is adapt for the result of one input item; item is the item's
// driver value (e.g. the --id), available as {id} in --outputFileRaw.
func (r *Runner) adaptItem(v any, item string) output.Seq {
	var seq output.Seq
	switch x := v.(type) {
	case nil:
		return emptySeq
	case op.Result[core.BinaryResponse]:
		seq = r.binary(x, item)
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

// binary writes a downloaded file to --outputFileRaw (templated with {id} and
// {filename}) or streams it to stdout, yielding no documents. A dry run or
// deferred result was never received, so nothing is written.
func (r *Runner) binary(res op.Result[core.BinaryResponse], item string) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if res.Err != nil {
			yield(jsondoc.Empty(), res.Err)
			return
		}
		if res.Request != nil || res.IsDeferred() {
			return
		}
		resp := res.Data
		defer resp.Close()
		if path := r.Config.GetOutputFileRaw(); path != "" {
			fields := map[string]string{"id": item, "filename": resp.FileName()}
			if _, err := fileutilities.WriteToFile(resp.Reader(), path, false, false, fields); err != nil {
				yield(jsondoc.Empty(), err)
			}
			return
		}
		if _, err := io.Copy(r.Cmd.OutOrStdout(), resp.Reader()); err != nil {
			yield(jsondoc.Empty(), err)
		}
	}
}
