package c8ystream

import (
	"bytes"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/tidwall/gjson"
)

// rawFileStage tees each document's raw bytes to the --outputFileRaw path. It
// runs early — right after error collection, before --filter/--select/output
// template — so the file holds the unshaped response, the way the v1 request
// handler wrote resp.Body() before any client-side processing.
//
// The path may contain {id}/{name}/{type}/{owner} placeholders, substituted per
// document, so a stream writes one file per item (e.g. "out/{id}.json"). Writes
// truncate (not append), matching v1: when several documents resolve to the
// same path the last one wins.
func rawFileStage(pathTemplate string) output.Stage {
	return Tap(func(doc jsondoc.JSONDoc) error {
		raw := doc.Raw()
		fields := map[string]string{
			"id":    gjson.GetBytes(raw, "id").String(),
			"name":  gjson.GetBytes(raw, "name").String(),
			"type":  gjson.GetBytes(raw, "type").String(),
			"owner": gjson.GetBytes(raw, "owner").String(),
		}
		_, err := fileutilities.WriteToFile(bytes.NewReader(raw), pathTemplate, false, true, fields)
		return err
	})
}
