package repodetails

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
)

func TestVersionWithoutParent(t *testing.T) {
	var mu sync.Mutex
	var requests []string
	empty := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/currentTenant" {
			mu.Lock()
			requests = append(requests, r.URL.RequestURI())
			mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		if empty {
			_, _ = w.Write([]byte(`{"managedObjects":[],"statistics":{"pageSize":5,"currentPage":1}}`))
			return
		}
		_, _ = w.Write([]byte(`{"managedObjects":[{"id":"501","type":"c8y_SoftwareBinary"},{"id":"502"}],"statistics":{"pageSize":5,"currentPage":1}}`))
	}))
	defer srv.Close()
	client := apiv2.NewClient(apiv2.ClientOptions{BaseURL: srv.URL})
	ctx := context.Background()

	for _, id := range []string{"", "12345"} {
		if got, err := VersionWithoutParent(ctx, client, SoftwareVersionKind, id); err != nil || got != id {
			t.Errorf("%q is used as is, got %q, %v", id, got, err)
		}
	}
	if len(requests) != 0 {
		t.Fatalf("no request for an id, got %v", requests)
	}

	tests := []struct {
		kind    VersionKind
		wants   []string
		wantNot string
	}{
		{SoftwareVersionKind, []string{"c8y_SoftwareBinary", "not%28has%28c8y_Patch%29%29", "c8y_Software.version+eq+%271.0.0%27", "pageSize=5"}, ""},
		{FirmwareVersionKind, []string{"c8y_FirmwareBinary", "not%28has%28c8y_Patch%29%29", "c8y_Firmware.version+eq+%271.0.0%27"}, ""},
		{FirmwarePatchKind, []string{"c8y_FirmwareBinary", "has%28c8y_Patch%29", "c8y_Firmware.version+eq+%271.0.0%27"}, "not%28has"},
	}
	for _, tc := range tests {
		requests = nil
		got, err := VersionWithoutParent(ctx, client, tc.kind, "1.0.0")
		if err != nil || got != "501" {
			t.Fatalf("%v: got %q, %v, want the first match", tc.kind, got, err)
		}
		if len(requests) != 1 {
			t.Fatalf("want 1 request (the version is not looked up by its software), got %v", requests)
		}
		for _, want := range tc.wants {
			if !strings.Contains(requests[0], want) {
				t.Errorf("%v: request %q has no %q", tc.kind, requests[0], want)
			}
		}
		if tc.wantNot != "" && strings.Contains(requests[0], tc.wantNot) {
			t.Errorf("%v: request %q has %q", tc.kind, requests[0], tc.wantNot)
		}
		if strings.Contains(requests[0], "bygroupid") {
			t.Errorf("%v: request %q is scoped to a parent", tc.kind, requests[0])
		}
	}

	empty = true
	_, err := VersionWithoutParent(ctx, client, SoftwareVersionKind, "9.9.9")
	if err == nil || !strings.Contains(err.Error(), "no matching items found. name=9.9.9") {
		t.Errorf("an unknown version is an unresolved name, got %v", err)
	}

	_, err = VersionWithoutParent(ctx, apiv2.NewClient(apiv2.ClientOptions{BaseURL: "http://127.0.0.1:1"}), SoftwareVersionKind, "1.0.0")
	var lookupErr *core.LookupError
	if !errors.As(err, &lookupErr) {
		t.Errorf("want a lookup error, got %v", err)
	}
}
