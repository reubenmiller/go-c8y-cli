package c8ystream

import (
	"reflect"
	"testing"
)

func TestLeafPaths(t *testing.T) {
	tests := []struct {
		name     string
		doc      string
		selected []string
		want     []string
	}{
		{"simple array is one column", `{"name":"a","tags":["x","y"]}`, nil, []string{"name", "tags"}},
		{"empty array is one column", `{"tags":[]}`, nil, []string{"tags"}},
		{"array of objects has a column per index", `{"items":[{"id":"1"},{"id":"2"}]}`, nil, []string{"items.0.id", "items.1.id"}},
		{"selected element", `{"name":"a","tags":[null,"y"]}`, []string{"name", "tags.1"}, []string{"name", "tags.1"}},
		{"nested objects", `{"a":{"b":1,"c":{"d":2}}}`, nil, []string{"a.b", "a.c.d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := leafPaths([]byte(tt.doc), tt.selected)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("leafPaths = %v, want %v", got, tt.want)
			}
		})
	}
}
