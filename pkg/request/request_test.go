package request

import (
	"net/url"
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
