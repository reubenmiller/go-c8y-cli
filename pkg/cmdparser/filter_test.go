package cmdparser

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/filter"
)

var analyticsModels = []string{
	`{"id":"1","name":"model A"}`,
	`{"id":"2","name":"template A","c8y_Template":{"version":"1"}}`,
	`{"id":"3","name":"model B"}`,
}

// TestCommandFilter checks that the filters of an extension command use the
// --filter syntax which the runner applies (c8ystream.Runner.Filter), and
// that they are combined with the user's filters.
func TestCommandFilter(t *testing.T) {
	testcases := []struct {
		name       string
		spec       []string
		userFilter []string
		expected   []string
	}{
		{name: "command filter only", spec: []string{"has c8y_Template"}, expected: []string{"2"}},
		{name: "command filter with nothas operator", spec: []string{"nothas c8y_Template"}, expected: []string{"1", "3"}},
		{name: "combined with user filter", spec: []string{"nothas c8y_Template"}, userFilter: []string{"name like *B"}, expected: []string{"3"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pred, err := filter.Parse(append(tc.spec, tc.userFilter...)...)
			if err != nil {
				t.Fatal(err)
			}
			got := []string{}
			for _, item := range analyticsModels {
				doc := jsondoc.New([]byte(item))
				if pred(doc) {
					got = append(got, doc.Get("id").String())
				}
			}
			if strings.Join(got, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("unexpected results. got=%v, want=%v", got, tc.expected)
			}
		})
	}
}

func TestParseSpecWithFilter(t *testing.T) {
	spec, err := ParseSpec(strings.NewReader(`
group:
  name: analytics
commands:
  - name: list
    method: GET
    path: /service/cep/analyticsbuilder
    collectionProperty: analyticsBuilderModelRepresentations
    filter:
      - nothas c8y_Template
      - name like "*demo*"
`))
	if err != nil {
		t.Fatal(err)
	}
	filter := spec.Commands[0].Filter
	if len(filter) != 2 || filter[1] != `name like "*demo*"` {
		t.Errorf("unexpected filter. got=%v", filter)
	}
}
