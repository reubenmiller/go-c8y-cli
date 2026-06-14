package c8ystream

import (
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/tidwall/gjson"
	"resty.dev/v3"
)

// installActivityLog registers a resty response middleware that writes one
// activity-log entry per request/response, the way the v1 request handler did.
// Doing it at the transport captures every request the command makes —
// pagination pages and name → id resolution lookups included — uniformly,
// rather than only the documents that reach the output pipeline.
//
// The activity logger itself filters by method and the enabled flag, so this is
// a no-op when logging is off. Dry-run and deferred-prepare responses (which
// carry the X-Dry-Run header and were never actually sent) are skipped so the
// log reflects real traffic only.
func (r *Runner) installActivityLog(client *apiv2.Client) error {
	al, err := r.Factory.ActivityLogger()
	if err != nil {
		return err
	}
	if al == nil || client == nil || client.HTTPClient == nil {
		return nil
	}
	client.HTTPClient.AddResponseMiddleware(func(_ *resty.Client, resp *resty.Response) error {
		if resp == nil || resp.RawResponse == nil {
			return nil
		}
		if resp.RawResponse.Header.Get("X-Dry-Run") != "" {
			return nil
		}
		al.LogRequest(resp.RawResponse, gjson.ParseBytes(resp.Bytes()), resp.Duration().Milliseconds())
		return nil
	})
	return nil
}
