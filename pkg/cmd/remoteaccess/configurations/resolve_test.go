package configurations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// TestResolveConfigurationID checks that a name which the device has no
// configuration of is reported as an unresolved name (exit code 4), found by the
// type of the error of the SDK, not by its text.
func TestResolveConfigurationID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","name":"webssh","protocol":"SSH"}]`))
	}))
	defer srv.Close()
	client := apiv2.NewClient(apiv2.ClientOptions{BaseURL: srv.URL})
	ctx := context.Background()

	if id, err := resolveConfigurationID(ctx, client, "d1", "WebSSH"); err != nil || id != "1" {
		t.Errorf("by name: got %q, %v", id, err)
	}
	if id, err := resolveConfigurationID(ctx, client, "d1", "42"); err != nil || id != "42" {
		t.Errorf("by id: got %q, %v", id, err)
	}
	_, err := resolveConfigurationID(ctx, client, "d1", "missing")
	ce, ok := err.(cmderrors.CommandError)
	if !ok || ce.ExitCode != cmderrors.ExitNotFound404 || ce.Error() != "commandError: referenceByName: no matching items found. name=missing  ::StatusCode=404" {
		t.Errorf("unknown name: got %#v", err)
	}
}
