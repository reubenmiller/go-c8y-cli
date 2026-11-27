package c8ystream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinDocs(t *testing.T) {
	assert.Equal(t, "null", string(joinDocs(nil)))
	assert.Equal(t, `{"id":"1"}`, string(joinDocs([][]byte{[]byte(`{"id":"1"}`)})))
	assert.Equal(t, `[{"id":"1"},{"id":"2"}]`, string(joinDocs([][]byte{[]byte(`{"id":"1"}`), []byte(`{"id":"2"}`)})))
}

// a list gives the template the array of its items, also for one item or none (v2.59.0)
func TestTemplateBodiesOfAList(t *testing.T) {
	r := &Runner{}
	for name, tt := range map[string]struct {
		pages []string
		docs  []string
		want  []string
	}{
		"empty list":     {[]string{`{"managedObjects":[],"next":"n"}`}, nil, []string{`[]`}},
		"one item":       {[]string{`{"managedObjects":[{"id":"1"}]}`}, []string{`{"id":"1"}`}, []string{`[{"id":"1"}]`}},
		"two pages":      {[]string{`{"managedObjects":[{"id":"1"}]}`, `{"managedObjects":[{"id":"2"}]}`}, []string{`{"id":"1"}`, `{"id":"2"}`}, []string{`[{"id":"1"}]`, `[{"id":"2"}]`}},
		"a bare array":   {[]string{`[{"id":"1"}]`}, []string{`{"id":"1"}`}, []string{`[{"id":"1"}]`}},
		"not a list":     {[]string{`{"id":"1","name":"a"}`}, []string{`{"id":"1","name":"a"}`}, []string{`{"id":"1","name":"a"}`}},
		"command drops":  {[]string{`{"managedObjects":[{"id":"1"},{"id":"2"}]}`}, []string{`{"id":"1"}`}, []string{`{"id":"1"}`}},
		"nothing to see": {nil, nil, []string{`null`}},
	} {
		t.Run(name, func(t *testing.T) {
			r.tmplPages = nil
			for _, p := range tt.pages {
				r.tmplPages = append(r.tmplPages, templatePage{method: "GET", path: "/inventory/managedObjects", body: []byte(p)})
			}
			docs := make([][]byte, 0, len(tt.docs))
			for _, d := range tt.docs {
				docs = append(docs, []byte(d))
			}
			got := []string{}
			for _, b := range r.templateBodies(docs, false) {
				got = append(got, string(b))
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// with --raw, --withTotalPages and --withTotalElements the template gets the whole response
func TestTemplateBodiesOfTheWholeResponse(t *testing.T) {
	r := &Runner{}
	r.tmplPages = []templatePage{{method: "GET", path: "/inventory/managedObjects", body: []byte(`{"managedObjects":[{"id":"1"}],"statistics":{"pageSize":5}}`)}}
	got := r.templateBodies([][]byte{[]byte(`{"id":"1"}`)}, true)
	assert.Equal(t, [][]byte{[]byte(`{"managedObjects":[{"id":"1"}],"statistics":{"pageSize":5}}`)}, got)
}
