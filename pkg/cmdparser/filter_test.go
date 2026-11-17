package cmdparser

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/tidwall/gjson"
)

const analyticsModels = `{"models":[
	{"id":"1","name":"model A"},
	{"id":"2","name":"template A","c8y_Template":{"version":"1"}},
	{"id":"3","name":"model B"}
]}`

func filteredIDs(t *testing.T, filters *jsonfilter.JSONFilters) []string {
	t.Helper()
	out, err := filters.Apply(analyticsModels, "models", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, item := range gjson.ParseBytes(out).Array() {
		ids = append(ids, item.Get("id").String())
	}
	return ids
}

func TestApplyCommandFilter(t *testing.T) {
	testcases := []struct {
		name       string
		spec       []string
		userFilter []string
		expected   []string
	}{
		{name: "no filters", expected: []string{"1", "2", "3"}},
		{name: "command filter only", spec: []string{"has c8y_Template"}, expected: []string{"2"}},
		{name: "command filter with nothas operator", spec: []string{"nothas c8y_Template"}, expected: []string{"1", "3"}},
		{name: "combined with user filter", spec: []string{"nothas c8y_Template"}, userFilter: []string{"name like *B"}, expected: []string{"3"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			filters := jsonfilter.NewJSONFilters(logger.NewDummyLogger("test"))
			if err := filters.AddRawFilters(tc.userFilter); err != nil {
				t.Fatal(err)
			}
			options := config.CommonCommandOptions{Filters: filters}
			if err := ApplyCommandFilter(options, models.Command{Name: "list", Filter: tc.spec}); err != nil {
				t.Fatal(err)
			}
			got := filteredIDs(t, options.Filters)
			if strings.Join(got, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("unexpected results. got=%v, want=%v", got, tc.expected)
			}
		})
	}
}

func TestApplyCommandFilterWithoutFilters(t *testing.T) {
	if err := ApplyCommandFilter(config.CommonCommandOptions{}, models.Command{Filter: []string{"name eq a"}}); err != nil {
		t.Errorf("expected no error when there are no output filters. got=%s", err)
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
