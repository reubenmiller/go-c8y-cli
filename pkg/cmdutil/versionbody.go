package cmdutil

import (
	"encoding/json"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/tidwall/gjson"
)

// VersionBody is the part of a --template/--data body which the software and
// firmware version/patch create commands take over: the SDK builds the
// version document itself from a version and a url, so the values a template
// (or piped input) sets in the typed fragment are read back from it.
type VersionBody struct {
	// Version and URL are the template values of <fragment>.version and
	// <fragment>.url (empty when the template does not set them)
	Version string
	URL     string

	// Annotations are the other top-level fragments of the body, to be written
	// into the version document as they are
	Annotations []model.Fragment
}

// ParseVersionBody reads the body which the version create commands built from
// --data/--template (e.g. {"c8y_Software":{"version":"1.0.0"}}). fragment is the
// key holding the version and url ("c8y_Software" or "c8y_Firmware").
func ParseVersionBody(body json.RawMessage, fragment string) VersionBody {
	out := VersionBody{}
	if len(body) == 0 {
		return out
	}
	doc := gjson.ParseBytes(body)
	if !doc.IsObject() {
		return out
	}
	out.Version = doc.Get(fragment + ".version").String()
	out.URL = doc.Get(fragment + ".url").String()
	doc.ForEach(func(key, value gjson.Result) bool {
		if key.String() == fragment {
			return true
		}
		var v any
		if err := json.Unmarshal([]byte(value.Raw), &v); err == nil {
			out.Annotations = append(out.Annotations, model.Frag(key.String(), v))
		}
		return true
	})
	return out
}
