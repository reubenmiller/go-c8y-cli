package c8ystream

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
)

// recordingClient is a client of a server which answers with the given body and
// records the requests (without the tenant discovery of the client).
func recordingClient(t *testing.T, body string) (*apiv2.Client, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/currentTenant" {
			mu.Lock()
			requests = append(requests, r.Method+" "+r.URL.RequestURI())
			mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return apiv2.NewClient(apiv2.ClientOptions{BaseURL: srv.URL}), func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), requests...)
	}
}

func unreachableClient() *apiv2.Client {
	return apiv2.NewClient(apiv2.ClientOptions{BaseURL: "http://127.0.0.1:1"})
}

func TestResolveHostedApplication(t *testing.T) {
	client, requests := recordingClient(t, `{"applications":[{"id":"11","name":"cockpit","type":"HOSTED"},{"id":"12","name":"cockpit2","type":"HOSTED"}],"statistics":{"pageSize":2000,"currentPage":1}}`)
	ctx := context.Background()

	if id, err := ResolveHostedApplication(ctx, client, "t1", "123"); err != nil || id != "123" {
		t.Errorf("an id is used as is: got %q, %v", id, err)
	}
	if got := requests(); len(got) != 0 {
		t.Errorf("an id needs no request, got %v", got)
	}
	if id, err := ResolveHostedApplication(ctx, client, "t1", "cockpit"); err != nil || id != "11" {
		t.Errorf("by name: got %q, %v", id, err)
	}
	// the name must match exactly
	if id, err := ResolveHostedApplication(ctx, client, "t1", "cock*"); err != nil || id != "" {
		t.Errorf("wildcards are not expanded: got %q, %v", id, err)
	}
	// the applications of the tenant are asked first, then all of them
	if id, _ := ResolveHostedApplication(ctx, client, "t1", "nosuch"); id != "" {
		t.Errorf("unknown name: got %q", id)
	}
	got := requests()
	want := "GET /application/applications?owner=t1&pageSize=2000&type=HOSTED"
	if got[0] != want {
		t.Errorf("first request = %q, want %q", got[0], want)
	}

	if _, err := HostedApplicationRef(ctx, client, "t1", "nosuch"); err == nil || !strings.Contains(err.Error(), "no matching items found. name=nosuch") {
		t.Errorf("an unknown name is an unresolved name, got %v", err)
	}
	if id, err := HostedApplicationRef(ctx, client, "t1", ""); err != nil || id != "" {
		t.Errorf("an empty value stays empty: got %q, %v", id, err)
	}

	// a request which fails is a lookup error
	_, err := ResolveHostedApplication(ctx, unreachableClient(), "t1", "cockpit")
	var lookupErr *core.LookupError
	if !errors.As(err, &lookupErr) || lookupErr.Resource != "hosted application" {
		t.Errorf("want a lookup error of a hosted application, got %v", err)
	}
}

func TestUserWithWildcards(t *testing.T) {
	client, requests := recordingClient(t, `{"users":[{"id":"admin","userName":"admin"},{"id":"adm-other","userName":"adm-other"}],"statistics":{"pageSize":5,"currentPage":1}}`)
	ctx := context.Background()

	if got, err := UserWithWildcards(ctx, client, "admin"); err != nil || got != "admin" {
		t.Errorf("a name without wildcard is used as is: got %q, %v", got, err)
	}
	if len(requests()) != 0 {
		t.Errorf("no request for a name without wildcard, got %v", requests())
	}
	if got, err := UserWithWildcards(ctx, client, "adm*"); err != nil || got != "admin" {
		t.Errorf("a wildcard resolves to the first match: got %q, %v", got, err)
	}
	if got := requests(); len(got) != 1 || !strings.Contains(got[0], "pageSize=5") || !strings.Contains(got[0], "username=adm") {
		t.Errorf("v2.59.0 asked for a page of 5 users filtered by the name without wildcards, got %v", got)
	}
	_, err := UserWithWildcards(ctx, client, "x*y")
	var ce cmderrors.CommandError
	if !errors.As(err, &ce) || ce.ExitCode != cmderrors.ExitNotFound404 {
		t.Errorf("no match is an unresolved name (exit 4), got %v", err)
	}
	_, err = UserWithWildcards(ctx, unreachableClient(), "adm*")
	if !core.IsLookupError(err) {
		t.Errorf("want a lookup error, got %v", err)
	}
}
