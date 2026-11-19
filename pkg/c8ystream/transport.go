package c8ystream

import (
	"net/http"
	"sync"
)

// dryRunHandler returns the callback go-c8y invokes for each prepared request
// in dry-run mode. It reuses the legacy request handler's PrintRequestDetails
// so the output (and --dryFormat: json/dump/curl/markdown) matches the v1
// commands exactly, which the test suite depends on. Writes are serialised
// because calls may run on the worker pool.
func (r *Runner) dryRunHandler() (func(*http.Request), error) {
	handler, err := r.Factory.GetRequestHandler()
	if err != nil {
		return nil, err
	}
	out := r.Cmd.OutOrStdout()
	var mu sync.Mutex
	return func(req *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		// requestOptions is only used for curl generation, which reads the
		// request itself, so nil is safe here.
		handler.PrintRequestDetails(out, nil, req)
	}, nil
}
