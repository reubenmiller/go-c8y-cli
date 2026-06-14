package c8ystream

import (
	"context"
	"testing"

	"github.com/tidwall/gjson"
)

func TestResolveSourceID(t *testing.T) {
	in := &Resolver{ctx: context.Background()}

	var calls []string
	resolve := func(_ context.Context, ref string) (string, error) {
		calls = append(calls, ref)
		if ref == "name:myDevice" {
			return "98765", nil
		}
		return ref, nil // ids pass through
	}

	// 1. no source.id -> body unchanged, resolver not called
	calls = nil
	body, err := in.ResolveSourceID([]byte(`{"type":"x"}`), resolve)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(body) != `{"type":"x"}` {
		t.Errorf("expected unchanged body, got %s", body)
	}
	if len(calls) != 0 {
		t.Errorf("resolver should not be called when source.id absent, calls=%v", calls)
	}

	// 2. plain id -> passthrough (resolve sees the id, body keeps it)
	calls = nil
	body, err = in.ResolveSourceID([]byte(`{"source":{"id":"12345"},"type":"x"}`), resolve)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got := gjson.GetBytes(body, "source.id").String(); got != "12345" {
		t.Errorf("source.id = %q, want 12345", got)
	}
	if len(calls) != 1 || calls[0] != "12345" {
		t.Errorf("resolve calls = %v, want [12345]", calls)
	}

	// 3. name -> resolved to id and replaced in the body
	calls = nil
	body, err = in.ResolveSourceID([]byte(`{"source":{"id":"myDevice"},"type":"x"}`), resolve)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got := gjson.GetBytes(body, "source.id").String(); got != "98765" {
		t.Errorf("source.id = %q, want 98765 (resolved)", got)
	}
	if len(calls) != 1 || calls[0] != "name:myDevice" {
		t.Errorf("resolve calls = %v, want [name:myDevice]", calls)
	}
	// other fields preserved
	if got := gjson.GetBytes(body, "type").String(); got != "x" {
		t.Errorf("type = %q, want x", got)
	}
}
