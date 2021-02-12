package jsonformatter

import (
	"bytes"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

// ByteFormatter formats/transforms a byte array
type ByteFormatter func([]byte) []byte

func dummyFormatter(input []byte) []byte {
	return input
}

// WithOutputFormatters applies given formatters to byte array
func WithOutputFormatters(input []byte, out bool, formatters ...ByteFormatter) []byte {
	for _, formatter := range formatters {
		input = formatter(input)
	}

	if out {
		fmt.Printf("%s", input)
	}
	return input
}

// WithPrettyPrint formatts the json into more human friendly format
func WithPrettyPrint(enabled bool) ByteFormatter {
	return WithOptionalFormatter(enabled, pretty.Pretty)
}

func WithSuffix(enabled bool, suffix string) ByteFormatter {
	return WithOptionalFormatter(enabled, func(input []byte) []byte {
		if enabled {
			return append(input, []byte(suffix)...)
		}
		return input
	})
}

// WithTrimSpace formatts the json into more human friendly format
func WithTrimSpace(enabled bool) ByteFormatter {
	return WithOptionalFormatter(enabled, bytes.TrimSpace)
}

// WithJSONStreamOutput converts json to json lines so it can be processed in the pipe
func WithJSONStreamOutput(enabled bool, stream bool, compact bool) ByteFormatter {
	return WithOptionalFormatter(enabled, func(input []byte) []byte {
		formatter := dummyFormatter
		if !compact {
			formatter = pretty.Pretty
		}
		j := gjson.ParseBytes(input)
		if j.IsArray() {
			if stream {
				j.ForEach(func(key, value gjson.Result) bool {
					fmt.Printf("%s\n", formatter([]byte(value.Raw)))
					return true
				})
			} else {
				fmt.Printf("%s\n", formatter(input))
			}
		} else if j.IsObject() {
			fmt.Printf("%s\n", formatter(input))
		} else {
			// No formatter (as it could be non-json)
			if len(input) != 0 {
				fmt.Printf("%s\n", input)
			}
		}
		return input
	})
}

// WithOptionalFormatter formatter which is only applied if it is enabled
func WithOptionalFormatter(enabled bool, formatter ByteFormatter) ByteFormatter {
	return func(input []byte) []byte {
		if enabled {
			return formatter(input)
		}
		return input
	}
}
