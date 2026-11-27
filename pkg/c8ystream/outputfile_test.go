package c8ystream

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
)

func TestOutputFileJSONLines(t *testing.T) {
	var buf bytes.Buffer
	r := &jsonLinesRenderer{w: &noTrailingNewline{w: &buf}}
	for _, doc := range []string{`{"id": "1",  "name": "a<b"}`, `{"id":"2"}`, `text`} {
		if err := r.Write(jsondoc.New([]byte(doc))); err != nil {
			t.Fatal(err)
		}
	}
	// one compact document per line (v2.59.0), no newline after the last one
	want := "{\"id\":\"1\",\"name\":\"a<b\"}\n{\"id\":\"2\"}\ntext"
	if got := buf.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNoTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	w := &noTrailingNewline{w: &buf}
	for _, s := range []string{"a,b\n", "c,d\n", "\n"} {
		if _, err := w.Write([]byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	if got, want := buf.String(), "a,b\nc,d\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDropNoPageSizeHook(t *testing.T) {
	tests := []struct {
		name, method, query, want string
	}{
		{"default page size", http.MethodGet, "pageSize=1999", ""},
		{"with other parameters", http.MethodGet, "a=1&pageSize=1999&q=x", "a=1&q=x"},
		{"explicit page size", http.MethodGet, "pageSize=7&q=x", "pageSize=7&q=x"},
		{"page size 19990", http.MethodGet, "pageSize=19990", "pageSize=19990"},
		{"page size 2000", http.MethodGet, "pageSize=2000", "pageSize=2000"},
		{"no query", http.MethodGet, "", ""},
		{"not a GET", http.MethodPost, "pageSize=1999", "pageSize=1999"},
	}
	hook := dropNoPageSizeHook()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{Method: tt.method, URL: &url.URL{Path: "/x", RawQuery: tt.query}}
			if err := hook(req); err != nil {
				t.Fatal(err)
			}
			if req.URL.RawQuery != tt.want {
				t.Fatalf("got %q, want %q", req.URL.RawQuery, tt.want)
			}
		})
	}
}
