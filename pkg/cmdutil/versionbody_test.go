package cmdutil

import (
	"testing"
)

func TestParseVersionBody(t *testing.T) {
	body := []byte(`{"c8y_Software":{"version":"1.0.0","url":"https://example.com"},"custom":{"a":1},"type":"c8y_SoftwareBinary"}`)
	got := ParseVersionBody(body, "c8y_Software")
	if got.Version != "1.0.0" || got.URL != "https://example.com" {
		t.Errorf("version/url = %q/%q", got.Version, got.URL)
	}
	if len(got.Annotations) != 2 {
		t.Fatalf("annotations = %v", got.Annotations)
	}
	if got.Annotations[0].FragmentKey() != "custom" || got.Annotations[1].FragmentKey() != "type" {
		t.Errorf("annotations = %v", got.Annotations)
	}

	empty := ParseVersionBody(nil, "c8y_Software")
	if empty.Version != "" || empty.URL != "" || len(empty.Annotations) != 0 {
		t.Errorf("empty body = %+v", empty)
	}
}
