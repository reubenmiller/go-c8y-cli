package jsonformatter

import (
	"bytes"
	"fmt"
	"io"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/tidwall/gjson"
)

// OutputFormatter formats/transforms a byte array
type OutputFormatter func(io.Writer, []byte) []byte

// ByteFormatter formats/transforms a byte array
type ByteFormatter func([]byte) []byte

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

// WithSuffix adds an optional suffix to the output
func WithSuffix(enabled bool, suffix string) OutputFormatter {
	return WithOptionalFormatter(enabled, func(input []byte) []byte {
		if enabled {
			return append(input, []byte(suffix)...)
		}
		return input
	})
}

// WithTrimSpace formats the json into more human friendly format
func WithTrimSpace(enabled bool) OutputFormatter {
	return func(i io.Writer, b []byte) []byte {
		return WithOptionalFormatter(enabled, bytes.TrimSpace)(i, b)
	}
}

// WithJSONStreamOutput converts json to json lines so it can be processed in the pipe
func WithJSONStreamOutput(enabled bool, stream bool, isText bool) OutputFormatter {
	return func(i io.Writer, b []byte) []byte {
		return WithOptionalFormatter(enabled, func(input []byte) []byte {
			if isText {
				if len(input) > 0 {
					fmt.Fprintf(i, "%s\n", input)
				}
				return input
			}

			gjson.ForEachLine(string(input), func(line gjson.Result) bool {
				if line.IsArray() {
					if stream {
						line.ForEach(func(key, value gjson.Result) bool {
							fmt.Fprintf(i, "%s\n", []byte(value.Raw))
							return true
						})
					} else {
						fmt.Fprintf(i, "%s\n", []byte(line.Raw))
					}
				} else if line.IsObject() {
					fmt.Fprintf(i, "%s\n", []byte(line.Raw))
				} else {
					if len(line.Raw) > 0 {
						fmt.Fprintf(i, "%s\n", []byte(line.Raw))
					}
				}
				return true
			})
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

func writeToFile(text []byte, filename string, append bool) error {
	fields := make(map[string]string)

	if bytes.Contains(text, []byte("{")) && bytes.Contains(text, []byte("}")) {
		// Only works if it is json, otherwise these values will be empty
		props := []string{"id", "name", "type", "owner"}
		for i, result := range gjson.GetManyBytes(text, props...) {
			fields[props[i]] = result.String()
		}
	}

	_, err := fileutilities.WriteToFile(bytes.NewReader(text), filename, append, true, fields)

	if err != nil {
		return fmt.Errorf("failed to copy file contents to file. %s", err)
	}

	return nil
}

// WithFileOutput writes the response to file if enabled
func WithFileOutput(enabled bool, filename string, append bool) OutputFormatter {
	return func(i io.Writer, b []byte) []byte {
		return WithOptionalFormatter(enabled, func(input []byte) []byte {
			err := writeToFile(b, filename, append)
			if err != nil {
				panic(err)
			}
			return input
		})(i, b)
	}
}
