package jsonfilter

import (
	"testing"

	glob "github.com/obeattie/ohmyglob"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/assert"
)

func Test_SimpleGlobMatch(t *testing.T) {

	globstr := "id"
	pattern, err := glob.Compile(globstr, &glob.Options{
		Separator:    '.',
		MatchAtStart: true,
		MatchAtEnd:   true,
	})
	assert.OK(t, err)

	src := map[string]interface{}{
		"id":          "12345",
		"name":        "hello",
		"source.id":   "9876",
		"source.name": "world",
	}

	dst := make(map[string]interface{})
	_, err = filterFlatMap(src, dst, []glob.Glob{pattern}, []string{"id"})
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, dst, `{"id":"12345"}`)

}

func MustCompileGlob(p string) glob.Glob {
	pattern, err := glob.Compile(p, &glob.Options{
		Separator:    '.',
		MatchAtStart: true,
		MatchAtEnd:   true,
	})
	if err != nil {
		panic(err)
	}
	return pattern
}

func Test_SimpleInvertedGlobMatch(t *testing.T) {

	pattern1 := MustCompileGlob("!id")
	pattern2 := MustCompileGlob("name")
	match := pattern1.MatchString("id") && !pattern1.IsNegative()
	assert.False(t, match)

	src := map[string]interface{}{
		"id":          "12345",
		"Name":        "hello",
		"source.id":   "9876",
		"source.name": "world",
	}

	dst := make(map[string]interface{})
	keys, err := filterFlatMap(src, dst, []glob.Glob{pattern1, pattern2}, []string{"", ""})
	assert.OK(t, err)
	assert.EqualMarshalJSON(t, dst, `{"name":"hello"}`)
	assert.EqualMarshalJSON(t, keys, `["name"]`)
}
