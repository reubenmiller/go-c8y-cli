package c8ystream

import "testing"

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
