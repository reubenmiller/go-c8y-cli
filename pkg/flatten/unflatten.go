package flatten

import (
	"strings"

	"github.com/tidwall/sjson"
)

// Unflatten converts a map where the keys are dot-notation paths, and convert
// them back to a nested map.
// keys with a number, i.e. "path.0" will be converted to arrays {"path":[0]}
func Unflatten(data map[string]interface{}) (output []byte, err error) {
	output = []byte("{}")

	for k, v := range data {
		// strip key helper and for the key be treated as an object key
		// even it it is a number
		k = strings.ReplaceAll(k, KeyPrefix, ":")

		output, err = sjson.SetBytes(output, k, v)
		if err != nil {
			return nil, err
		}
	}
	return
}
