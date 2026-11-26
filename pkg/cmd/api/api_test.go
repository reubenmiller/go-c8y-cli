package api

import (
	"net/url"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestURL(t *testing.T) {
	cases := []struct {
		template, item, prefix string
		host                   string
		path                   string
		query                  url.Values
	}{
		{"{url}", "/inventory/managedObjects?pageSize=1&type=a&type=b", "", "", "/inventory/managedObjects", url.Values{"pageSize": {"1"}, "type": {"a", "b"}}},
		{"/test/{url}/endpoint", "12345", "", "", "/test/12345/endpoint", url.Values{}},
		{"/endpoint?id={url}", "1", "", "", "/endpoint", url.Values{"id": {"1"}}},
		{"alarm/alarms", "", "", "", "/alarm/alarms", url.Values{}},
		{"/inventory", "", "/service", "", "/service/inventory", url.Values{}},
		{"/service/example/foo/", "", "", "", "/service/example/foo/", url.Values{}},
		// an absolute url is requested from its own host (v2.59.0), whether it is
		// the url argument (the template) or a piped item
		{"{url}", "https://t1.example.com/inventory/managedObjects/1?withParents=true", "", "https://t1.example.com", "/inventory/managedObjects/1", url.Values{"withParents": {"true"}}},
		{"http://other-host:8080/path/x?a=1", "", "", "http://other-host:8080", "/path/x", url.Values{"a": {"1"}}},
		{"http://other-host/path", "", "/service", "http://other-host", "/path", url.Values{}},
	}
	for _, c := range cases {
		host, path, query, err := requestURL(c.template, c.item, c.prefix)
		require.NoError(t, err)
		assert.Equal(t, c.host, host, c.template)
		assert.Equal(t, c.path, path, c.template)
		assert.Equal(t, c.query, query, c.template)
	}
}

func TestPipedURL(t *testing.T) {
	opts := &flags.PipelineOptions{Property: "url", Aliases: []string{"c8y_Firmware.url", "self"}}
	assert.Equal(t, "/a", pipedURL([]byte("/a\n"), opts))
	assert.Equal(t, "12345", pipedURL([]byte(`{"url":"12345","self":"x"}`), opts))
	assert.Equal(t, "https://h/fw", pipedURL([]byte(`{"c8y_Firmware":{"url":"https://h/fw"}}`), opts))
	assert.Equal(t, "", pipedURL([]byte(`{"name":"x"}`), opts))
}
