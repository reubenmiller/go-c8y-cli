package c8ystream

import (
	"testing"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// the escapes are built from parts: the backslash must not start an escape in this source
const (
	lt  = "\\" + "u003c"
	gt  = "\\" + "u003e"
	amp = "\\" + "u0026"
)

func TestUnescapeHTML(t *testing.T) {
	tests := map[string]struct{ in, want string }{
		"no escapes":      {`{"name":"a"}`, `{"name":"a"}`},
		"all three":       {`{"name":"a` + lt + `b` + gt + `c` + amp + `d"}`, `{"name":"a<b>c&d"}`},
		"url with &":      {`{"next":"http://x?a=1` + amp + `b=2"}`, `{"next":"http://x?a=1&b=2"}`},
		"other escapes":   {`{"name":"` + "\\" + `u00e4` + `"}`, `{"name":"` + "\\" + `u00e4` + `"}`},
		"escaped newline": {`{"a":"x\\ny"}`, `{"a":"x\\ny"}`},
		"plain text":      {"a" + amp + "b", "a&b"},
		// the blunt replacement of v2.59.0 would write the invalid JSON "a\<b"
		"escaped backslash":                  {`{"a":"a\\` + "u003c" + `b","b":"` + lt + `"}`, `{"a":"a\\` + "u003c" + `b","b":"<"}`},
		"escaped backslash before an escape": {`{"a":"a\\` + lt + `"}`, `{"a":"a\\<"}`},
		"escaped backslash in plain text":    {"a\\\\" + "u003c", "a\\<"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(unescapeHTML([]byte(tt.in))))
		})
	}
}

func TestUnescapeHTMLStage(t *testing.T) {
	src := output.Seq(func(yield func(jsondoc.JSONDoc, error) bool) {
		_ = yield(jsondoc.New([]byte(`{"name":"a`+amp+`b`+lt+`c"}`)), nil) && yield(jsondoc.New([]byte(`{"name":"plain"}`)), nil)
	})
	var got []string
	for doc, err := range unescapeHTMLStage()(src) {
		require.NoError(t, err)
		got = append(got, string(doc.Raw()))
	}
	assert.Equal(t, []string{`{"name":"a&b<c"}`, `{"name":"plain"}`}, got)
}
