package extsdk

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
)

// Print writes v to stdout as json. A slice or array is written as one json value per line,
// which c8y processes for commands declared with "output: json" in extension.yaml
func Print(v any) error {
	return Fprint(os.Stdout, v)
}

// Fprint writes v to w as json. A slice or array is written as one json value per line
func Fprint(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	rv := reflect.ValueOf(v)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		if _, isRaw := v.(json.RawMessage); !isRaw && rv.Type().Elem().Kind() != reflect.Uint8 {
			for i := 0; i < rv.Len(); i++ {
				if err := enc.Encode(rv.Index(i).Interface()); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return enc.Encode(v)
}
