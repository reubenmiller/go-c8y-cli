package c8ystream

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

func TestCollectionProperty(t *testing.T) {
	cases := map[string]string{
		`{"managedObjects":[{"id":"1"}],"next":"x","statistics":{}}`: "managedObjects",
		`{"id":"1","childDevices":{"references":[]}}`:                "",
		`{"id":"1","list":[]}`:                                       "",
		`{"a":[],"b":[]}`:                                            "",
		`{"name":"no array"}`:                                        "",
		`[{"id":"1"}]`:                                               "",
		`{"items":[],"k1":1,"k2":1,"k3":1,"k4":1,"k5":1,"k6":1,"k7":1,"k8":1,"k9":1,"k10":1}`: "",
	}
	for body, want := range cases {
		if got := collectionProperty([]byte(body)); got != want {
			t.Errorf("collectionProperty(%s) = %q, want %q", body, got, want)
		}
	}
}

func TestIsJSONDocument(t *testing.T) {
	cases := map[string]bool{
		`{"a":1}`:       true,
		` [1,2] `:       true,
		`"text"`:        false,
		`41`:            false,
		`hello world`:   false,
		`{broken`:       false,
		`<html></html>`: false,
	}
	for body, want := range cases {
		if got := isJSONDocument([]byte(body)); got != want {
			t.Errorf("isJSONDocument(%s) = %v, want %v", body, got, want)
		}
	}
}

// TestResultProperty checks that the command's collection property is used
// when the response has it, "-" disables unwrapping, and otherwise the
// collection property is detected.
func TestResultProperty(t *testing.T) {
	body := []byte(`{"references":[{"id":"1"}],"other":[],"self":"x"}`)
	cases := []struct {
		annotation string
		want       string
	}{
		{annotation: "", want: ""},
		{annotation: "references", want: "references"},
		{annotation: "missing", want: ""},
		{annotation: "-", want: ""},
	}
	for _, tc := range cases {
		cmd := &cobra.Command{}
		if tc.annotation != "" {
			flags.WithCollectionProperty(tc.annotation)(cmd)
		}
		r := &Runner{Cmd: cmd}
		if got := r.resultProperty(body); got != tc.want {
			t.Errorf("resultProperty(annotation=%q) = %q, want %q", tc.annotation, got, tc.want)
		}
	}

	r := &Runner{Cmd: &cobra.Command{}}
	if got := r.resultProperty([]byte(`{"managedObjects":[],"next":"x"}`)); got != "managedObjects" {
		t.Errorf("detected property = %q, want managedObjects", got)
	}
}

// TestFilter checks that the command's own filters are validated.
func TestFilter(t *testing.T) {
	r := &Runner{}
	if err := r.Filter(); err != nil || len(r.filters) != 0 {
		t.Errorf("no filters. err=%v, filters=%v", err, r.filters)
	}
	if err := r.Filter("nothas c8y_Template", "name like *B"); err != nil {
		t.Fatal(err)
	}
	if len(r.filters) != 2 {
		t.Errorf("filters = %v", r.filters)
	}
	if err := r.Filter("name like"); err == nil {
		t.Error("expected an error for an invalid filter")
	}
}
