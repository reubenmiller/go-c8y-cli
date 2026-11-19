package c8ystream

import (
	"fmt"
	"net/http"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// installSilentSuccessMessage registers a resty response middleware that gives
// the user feedback for "silent" successful responses — a 204, or a 2xx/3xx
// where either the response carries no Content-Type or the request sent no
// Accept header (e.g. bulk updates like `alarms updateCollection`, whose 200
// has nothing to render). It reproduces v1's confirmation line
// (pkg/request/request.go):
//
//	✓ Updated /alarm/alarms => 200 OK
//
// printed to stderr only in a terminal (or with settings.defaults.forceTTY),
// without a progress bar, and not when an output template shapes the output.
// Dry-run responses never reach response middlewares, so dry runs stay silent.
func (r *Runner) installSilentSuccessMessage(client *apiv2.Client) {
	if client == nil {
		return
	}
	if r.Config.GetOutputTemplate() != "" {
		return
	}
	if !r.Config.ForceTTY() && !(r.Factory.IOStreams.IsStdoutTTY() && !r.Config.ShowProgress()) {
		return
	}
	client.Use(observeResponses(func(o observedResponse) {
		raw := o.Response
		if raw.Request == nil {
			return
		}
		code := raw.StatusCode
		silent := code == http.StatusNoContent ||
			((raw.Header.Get("Content-Type") == "" || raw.Request.Header.Get("Accept") == "") && code >= 200 && code < 400)
		if !silent {
			return
		}

		cs := r.Factory.IOStreams.ColorScheme()
		actionColor := cs.Green
		actionText := ""
		switch raw.Request.Method {
		case http.MethodDelete:
			actionText = "Deleted"
			actionColor = cs.Red
		case http.MethodPut:
			actionText = "Updated"
		case http.MethodPost:
			actionText = "Created"
		case http.MethodPatch:
			actionText = "Patched"
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			// reads have no side effect worth confirming (v1 reached this path
			// via its default branch, but silent GETs are resolution lookups here)
			return
		default:
			actionText = raw.Request.Method
		}
		fmt.Fprintf(r.Factory.IOStreams.ErrOut, "%s %s %s => %s\n", cs.SuccessIconWithColor(actionColor), actionText, raw.Request.URL.Path, raw.Status)
	}))
}
