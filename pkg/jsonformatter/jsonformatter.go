package jsonformatter

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

// OutputFormatter formats/transforms a byte array
type OutputFormatter func(io.Writer, []byte) []byte

// ByteFormatter formats/transforms a byte array
type ByteFormatter func([]byte) []byte

func dummyFormatter(input []byte) []byte {
	return input
}

// WithOutputFormatters applies given formatters to byte array
func WithOutputFormatters(w io.Writer, input []byte, out bool, formatters ...OutputFormatter) []byte {
	for _, formatter := range formatters {
		input = formatter(w, input)
	}

	if out {
		fmt.Fprintf(w, "%s", input)
	}
	return input
}

// WithPrettyPrint formatts the json into more human friendly format
func WithPrettyPrint(enabled bool) OutputFormatter {
	return func(i io.Writer, b []byte) []byte {
		return WithOptionalFormatter(enabled, pretty.Pretty)(i, b)
	}
}

func WithSuffix(enabled bool, suffix string) OutputFormatter {
	return WithOptionalFormatter(enabled, func(input []byte) []byte {
		if enabled {
			return append(input, []byte(suffix)...)
		}
		return input
	})
}

// WithTrimSpace formatts the json into more human friendly format
func WithTrimSpace(enabled bool) OutputFormatter {
	return func(i io.Writer, b []byte) []byte {
		return WithOptionalFormatter(enabled, bytes.TrimSpace)(i, b)
	}
}

// WithJSONStreamOutput converts json to json lines so it can be processed in the pipe
func WithJSONStreamOutput(enabled bool, stream bool, compact bool) OutputFormatter {
	return func(i io.Writer, b []byte) []byte {
		return WithOptionalFormatter(enabled, func(input []byte) []byte {
			formatter := dummyFormatter
			if !compact {
				formatter = pretty.Pretty
			}
			j := gjson.ParseBytes(input)
			if j.IsArray() {
				if stream {
					j.ForEach(func(key, value gjson.Result) bool {
						fmt.Fprintf(i, "%s\n", formatter([]byte(value.Raw)))
						return true
					})
				} else {
					fmt.Fprintf(i, "%s\n", formatter(input))
				}
			} else if j.IsObject() {
				fmt.Fprintf(i, "%s\n", formatter(input))
			} else {
				// No formatter (as it could be non-json)
				if len(input) != 0 {
					fmt.Fprintf(i, "%s\n", input)
				}
			}
			return input
		})(i, b)
	}
}

// WithOptionalFormatter formatter which is only applied if it is enabled
func WithOptionalFormatter(enabled bool, formatter ByteFormatter) OutputFormatter {
	return func(i io.Writer, input []byte) []byte {
		if enabled {
			return formatter(input)
		}
		return input
	}
}
