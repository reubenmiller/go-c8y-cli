package c8ystream

import (
	"net/http"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/requestutil"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// installOutputTemplateCapture registers a resty response middleware that
// records the most recent real request/response so the --outputTemplate stage
// can bind them as the `request` and `response` template variables (matching
// v1's request.ExecuteTemplate). It is only installed when an output template
// is set.
//
// Like the input.value binding, this relies on the serial pipeline (the
// default): the middleware fires while a job runs its service call, before the
// resulting documents flow through the output-template stage, so the captured
// request/response match the document being shaped. Name → id resolution and
// pagination lookups also fire the middleware; the last one before the document
// is rendered wins, which is the document's own (primary) request. Dry-run
// responses (X-Dry-Run) are skipped — they were never sent and the dry-run
// handler renders the request separately.
func (r *Runner) installOutputTemplateCapture(client *apiv2.Client) {
	if client == nil {
		return
	}
	if r.Config.GetOutputTemplate() == "" && r.Config.GetOutputFileRaw() == "" {
		return
	}
	client.Use(observeResponses(func(o observedResponse) {
		raw := o.Response

		requestData := templateRequestData(raw.Request)

		responseData := map[string]any{
			"statusCode":    raw.StatusCode,
			"status":        raw.Status,
			"duration":      o.Duration.Milliseconds(),
			"contentLength": raw.ContentLength,
			"contentType":   raw.Header.Get("Content-Type"),
			"header":        requestutil.FlattenArrayMap(raw.Header),
			"proto":         raw.Proto,
			"body":          string(o.Body),
		}

		page := templatePage{body: o.Body, fields: rawFileFields(raw, o.Body)}
		if req := raw.Request; req != nil && req.URL != nil {
			page.method, page.path = req.Method, req.URL.Path
			// the responses of a job are kept with the job, as the jobs of
			// --workers run at the same time
			if capture, ok := req.Context().Value(templateCaptureKey{}).(*templateCapture); ok {
				capture.add(page, requestData, responseData)
			}
		}

		r.tmplMu.Lock()
		r.tmplRequest = requestData
		r.tmplResponse = responseData
		r.tmplMu.Unlock()
	}))
}

// templateRequestData is the `request` template variable of a request.
func templateRequestData(req *http.Request) map[string]any {
	requestData := make(map[string]any)
	if req != nil && req.URL != nil {
		u := req.URL
		requestData["path"] = u.Path
		requestData["pathEncoded"] = strings.Replace(u.String(), u.Scheme+"://"+u.Host, "", 1)
		requestData["host"] = u.Host
		requestData["url"] = u.String()
		requestData["query"] = requestutil.TryUnescapeURL(u.RawQuery)
		requestData["queryParams"] = requestutil.FlattenArrayMap(u.Query())
		requestData["method"] = req.Method
	}
	return requestData
}

// captureDryRunRequest records a request which a dry run only prepared, for
// --outputTemplate: the template gets `request` as the request that would be
// sent and an empty `response`, as in v2.59.0 (status code 0, no body, no
// headers), as dry-run requests never reach the response middleware.
func (r *Runner) captureDryRunRequest(req *http.Request) {
	if r.Config.GetOutputTemplate() == "" || req == nil || req.URL == nil {
		return
	}
	requestData := templateRequestData(req)
	responseData := map[string]any{
		"statusCode":    0,
		"status":        "",
		"duration":      int64(0),
		"contentLength": int64(0),
		"contentType":   "",
		"header":        map[string]any{},
		"proto":         "",
		"body":          "",
	}
	if capture, ok := req.Context().Value(templateCaptureKey{}).(*templateCapture); ok {
		capture.add(templatePage{method: req.Method, path: req.URL.Path, dry: true}, requestData, responseData)
	}
	r.tmplMu.Lock()
	r.tmplRequest = requestData
	r.tmplResponse = responseData
	r.tmplMu.Unlock()
}
