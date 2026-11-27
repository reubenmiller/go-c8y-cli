package c8ystream

import (
	"bytes"
	"context"
	"net/http"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/tidwall/gjson"
)

// wrapRawFileJob writes the response of a job to the --outputFileRaw path, as
// received from the server and before any client-side processing (the way the
// v1 request handler wrote resp.Body()): the whole body of a list, with its
// collection property and statistics, and the last page of several. The
// responses are the ones which the output template capture records.
//
// A document which has no response of its own (a result which was not read
// from a JSON response) is written as it is.
//
// The path may contain {id}/{name}/{type}/{owner} placeholders, substituted
// from the response, so a stream of items writes one file per item (e.g.
// "out/{id}.json"). Writes truncate (not append), matching v1: when several
// responses resolve to the same path the last one wins.
func (r *Runner) wrapRawFileJob(pathTemplate string, job Job) Job {
	return func(ctx context.Context) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			capture, ok := ctx.Value(templateCaptureKey{}).(*templateCapture)
			if !ok {
				capture = &templateCapture{}
				ctx = context.WithValue(ctx, templateCaptureKey{}, capture)
			}
			failed := false
			job(ctx)(func(doc jsondoc.JSONDoc, err error) bool {
				if err != nil {
					failed = true
					return yield(doc, err)
				}
				if pages, _, _ := capture.get(); len(pages) == 0 {
					if werr := writeRawFile(pathTemplate, doc.Raw()); werr != nil {
						failed = true
						return yield(jsondoc.Empty(), werr)
					}
				}
				return yield(doc, nil)
			})
			if failed {
				return
			}
			if pages, _, _ := capture.get(); len(pages) > 0 {
				if err := writeRawFile(pathTemplate, pages[len(pages)-1].body); err != nil {
					yield(jsondoc.Empty(), err)
				}
			}
		}
	}
}

func writeRawFile(pathTemplate string, raw []byte) error {
	fields := map[string]string{
		"id":    gjson.GetBytes(raw, "id").String(),
		"name":  gjson.GetBytes(raw, "name").String(),
		"type":  gjson.GetBytes(raw, "type").String(),
		"owner": gjson.GetBytes(raw, "owner").String(),
	}
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

// fileReceivedFunc reports whether installOutputFileCapture saw a response.
func (r *Runner) fileReceivedFunc() bool { return r.fileReceived.Load() }
