package requestutil

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

// Test_RequestDetailsJSONFields locks in the dry-run json contract of v2.59.0:
// url, host, pathEncoded, path, query, method, headers, body, shell and
// powershell. Extra fields (e.g. a decoded queryParams map) break consumers
// which match the whole document.
func Test_RequestDetailsJSONFields(t *testing.T) {
	details := &RequestDetails{
		URL:         "u",
		Host:        "h",
		PathEncoded: "pe",
		Path:        "p",
		Query:       "q",
		Method:      "GET",
		Headers:     map[string]string{"a": "b"},
		Body:        "x",
		Shell:       "s",
		PowerShell:  "ps",
	}
	out, err := json.Marshal(details)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]any{}
	if err := json.Unmarshal(out, &fields); err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for k := range fields {
		got = append(got, k)
	}
	sort.Strings(got)
	want := []string{"body", "headers", "host", "method", "path", "pathEncoded", "powershell", "query", "shell", "url"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dry run json fields\n got: %v\nwant: %v", got, want)
	}

	// every field of the struct must be one of the v2.59.0 fields
	rt := reflect.TypeOf(*details)
	if rt.NumField() != len(want) {
		t.Errorf("RequestDetails has %d fields, want %d", rt.NumField(), len(want))
	}
}
