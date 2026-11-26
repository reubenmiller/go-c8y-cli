package c8ystream

import (
	"net/http"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/spf13/viper"
)

// TestRequestDefaults checks the headers every request gets, as in v2.59.0:
// Accept: application/json when the request has none, an Accept header which
// is set stays, --noAccept sends none, and the HTTP client's Accept-Encoding
// is removed (the transport then sends "gzip" itself), while one given with
// --header is kept.
func TestRequestDefaults(t *testing.T) {
	cases := []struct {
		name         string
		noAccept     bool
		header       http.Header
		wantAccept   string
		wantEncoding string
	}{
		{name: "adds Accept", header: http.Header{}, wantAccept: "application/json"},
		{name: "keeps Accept", header: http.Header{"Accept": {"text/csv"}}, wantAccept: "text/csv"},
		{name: "no Accept with noAccept", noAccept: true, header: http.Header{}, wantAccept: ""},
		{name: "removes Accept-Encoding", header: http.Header{"Accept-Encoding": {"gzip, deflate"}}, wantAccept: "application/json", wantEncoding: ""},
		{name: "keeps a given Accept-Encoding", header: http.Header{"Accept-Encoding": {"identity"}}, wantAccept: "application/json", wantEncoding: "identity"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			v.Set(config.SettingsIgnoreAcceptHeader, tc.noAccept)
			req, err := http.NewRequest(http.MethodGet, "http://localhost/x", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header = tc.header
			if err := requestDefaults(config.NewConfig(v))(req); err != nil {
				t.Fatal(err)
			}
			if got := req.Header.Get("Accept"); got != tc.wantAccept {
				t.Errorf("Accept = %q, want %q", got, tc.wantAccept)
			}
			if got := req.Header.Get("Accept-Encoding"); got != tc.wantEncoding {
				t.Errorf("Accept-Encoding = %q, want %q", got, tc.wantEncoding)
			}
		})
	}
}
