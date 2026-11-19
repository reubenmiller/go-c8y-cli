package request

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func Test_optimizeManagedObjectsURL(t *testing.T) {

	testInputs := [][]string{
		// [input query, lastID, expected Query]
		{"https://myserver.com?query=$filter=(has(test)) $orderby=name asc", "", "$filter=(_id gt 0 and (has(test))) $orderby=_id asc"},
		{"https://myserver.com?query=$filter=has(test) $orderby=name asc", "0", "$filter=(_id gt 0 and (has(test))) $orderby=_id asc"},
		{"https://myserver.com?query=$filter=has(test)", "0", "$filter=(_id gt 0 and (has(test))) $orderby=_id asc"},
		{"https://myserver.com?query=$filter=(has(test))", "0", "$filter=(_id gt 0 and (has(test))) $orderby=_id asc"},
		{"https://myserver.com?query=$filter=(has(test))", "1000000000000", "$filter=(_id gt 1000000000000 and (has(test))) $orderby=_id asc"},
	}

	for _, item := range testInputs {
		u, _ := url.Parse(item[0])
		u = optimizeManagedObjectsURL(u, item[1])
		expectedQuery := "query=" + url.QueryEscape(item[2])
		if u.RawQuery != expectedQuery {
			t.Errorf("Query does not match. wanted=%s, got=%s", expectedQuery, u.RawQuery)
		}
	}
}

// Test_RequestDetailsQueryParams locks in the dry-run json contract for the
// queryParams field: it is the decoded map form of the query string, a
// parameter repeated in the query is preserved as a multi-value array (rather
// than collapsed, which a map[string]string could not do), and the field is
// omitted when the request carries no query.
func Test_RequestDetailsQueryParams(t *testing.T) {
	// A request URL with a repeated parameter and a single-value parameter, the
	// shape the json dry-run reports (e.g. `alarms list --status A --status B`).
	req, err := http.NewRequest(http.MethodGet, "https://example.c8y.io/alarm/alarms?status=ACTIVE&status=ACKNOWLEDGED&pageSize=100", nil)
	if err != nil {
		t.Fatal(err)
	}

	details := &RequestDetails{
		Method:      req.Method,
		Path:        req.URL.Path,
		Query:       req.URL.RawQuery,
		QueryParams: req.URL.Query(),
	}

	out, err := json.Marshal(details)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)

	// Repeated parameter preserved as an ordered array, single value still an array.
	for _, want := range []string{
		`"queryParams":`,
		`"status":["ACTIVE","ACKNOWLEDGED"]`,
		`"pageSize":["100"]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("queryParams json missing %q\n got: %s", want, got)
		}
	}

	// Omitted entirely when there is no query string.
	noQuery, err := json.Marshal(&RequestDetails{Method: http.MethodDelete, Path: "/inventory/managedObjects/12345"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(noQuery), "queryParams") {
		t.Errorf("expected queryParams to be omitted when empty, got: %s", noQuery)
	}
}
