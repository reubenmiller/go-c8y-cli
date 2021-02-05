package flags

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/pkg/timestamp"
	"github.com/spf13/cobra"
)

// GetOption gets the value from a flag and returns the value which can be set accordingly
type GetOption func(cmd *cobra.Command) (name string, value interface{}, err error)

// WithQueryParameters returns a query parameter values given from command line arguments
func WithQueryParameters(cmd *cobra.Command, query url.Values, opts ...GetOption) (err error) {
	for _, opt := range opts {
		name, value, err := opt(cmd)
		if err != nil {
			return err
		}
		queryValue := fmt.Sprintf("%v", value)
		if name != "" && queryValue != "" {
			query.Add(name, url.QueryEscape(queryValue))
		}
	}
	return
}

// WithPathParameters returns a path parameter values given from command line arguments
func WithPathParameters(cmd *cobra.Command, path map[string]string, opts ...GetOption) (err error) {
	for _, opt := range opts {
		name, value, err := opt(cmd)
		if err != nil {
			return err
		}
		if name != "" {
			switch v := value.(type) {
			case []string:
				path[name] = fmt.Sprintf("%s", strings.Join(v, ","))

			case []int:
				path[name] = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")

			default:
				path[name] = fmt.Sprintf("%v", value)
			}
		}
	}
	return
}

// WithHeaders sets header values from command line arguments
func WithHeaders(cmd *cobra.Command, header http.Header, opts ...GetOption) (err error) {
	for _, opt := range opts {
		name, value, err := opt(cmd)
		if err != nil {
			return err
		}
		headerValue := fmt.Sprintf("%v", value)
		if name != "" {
			header.Add(name, headerValue)
		}
	}
	return
}

// WithBody returns a body from given command line arguments
func WithBody(cmd *cobra.Command, body *mapbuilder.MapBuilder, opts ...GetOption) (err error) {

	for _, opt := range opts {
		name, value, err := opt(cmd)
		if err != nil {
			return err
		}

		switch v := value.(type) {
		case string:
			// only set non-empty values by default
			if v != "" {
				err = body.Set(name, value)
			}
		case FilePath:
			if v != "" {
				body.SetFile(string(v))
			}
		case map[string]interface{}:
			if v != nil {
				if name != "" {
					body.Set(name, v)

				} else {
					body.SetMap(v)
				}
			}
		default:
			err = body.Set(name, value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// WithBoolValue adds a boolean value from cli arguments to a query parameter
func WithBoolValue(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {
		src, dst, format := UnpackGetterOptions("", opts...)
		if cmd.Flags().Changed(src) {
			value, err := cmd.Flags().GetBool(src)

			if format != "" {
				return dst, applyFormatter(format, value), err
			}
			return dst, value, err
		}
		return "", false, nil
	}
}

// WithStringValue adds a string value from cli arguments
func WithStringValue(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {

		src, dst, format := UnpackGetterOptions("%s", opts...)

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}
		return dst, applyFormatter(format, value), err
	}
}

// WithStringDefaultValue adds a string value from cli arguments
func WithStringDefaultValue(defaultValue string, opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {

		src, dst, format := UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(src) {
			if defaultValue != "" {
				return dst, defaultValue, nil
			}
			return "", defaultValue, nil
		}

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}
		return dst, applyFormatter(format, value), err
	}
}

// WithStringSliceValues adds a string slice from cli arguments
func WithStringSliceValues(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {

		src, dst, format := UnpackGetterOptions("%s", opts...)

		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			return dst, values, err
		}

		nonEmptyValues := make([]string, 0)
		if len(values) > 0 {
			for _, value := range values {
				value = applyFormatter(format, value)
				if value != "" {
					nonEmptyValues = append(nonEmptyValues, value)
				}
			}
		}

		return dst, nonEmptyValues, err
	}
}

// WithIntValue adds a integer (int) value from cli arguments
func WithIntValue(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)
		value, err := cmd.Flags().GetInt(src)
		return dst, value, err
	}
}

// WithFloatValue adds a float (float32) value from cli arguments
func WithFloatValue(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)
		value, err := cmd.Flags().GetFloat32(src)
		return dst, value, err
	}
}

// WithRelativeTimestamp adds a timestamp (string) value from cli arguments
func WithRelativeTimestamp(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)
		value, err := cmd.Flags().GetString(src)

		if err != nil {
			return dst, value, err
		}

		value, err = cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}

		// ignore empty values
		if value == "" {
			return "", value, err
		}

		ts, err := timestamp.TryGetTimestamp(value)

		if err != nil {
			return dst, ts, err
		}

		// decode %2B with original "+" (if required) and
		// let the calling function handle it
		return dst, timestamp.DecodeC8yTimestamp(ts), nil
	}
}

func applyFormatter(format string, value interface{}) string {
	if strings.Contains(format, "%") {
		return fmt.Sprintf(format, value)
	}
	// format is a fixed string
	return format
}

func UnpackGetterOptions(defaultFormat string, options ...string) (src string, dst string, formatter string) {
	formatter = defaultFormat

	if len(options) == 1 {
		src = options[0]
		dst = options[0]
	} else if len(options) == 2 {
		src = options[0]
		dst = options[1]
	} else if len(options) >= 3 {
		src = options[0]
		dst = options[1]
		formatter = options[2]
	}
	// if dst == "" {
	// 	dst = src
	// }
	return
}

// FilePath is a string representation of a file path
type FilePath string

// WithFilePath adds a file path from cli arguments
func WithFilePath(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {

		src, dst, _ := UnpackGetterOptions("%s", opts...)

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}

		return dst, FilePath(value), err
	}
}

func WithDataValue(opts ...string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {

		src, dst, _ := UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(FlagDataName) {
			return "", "", nil
		}

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}

		data := make(map[string]interface{})

		err = jsonUtilities.ParseJSON(resolveContents(value), data)
		if err != nil {
			return dst, "", fmt.Errorf("json error: %s parameter does not contain valid json or shorthand json. %w", src, err)
		}

		c8ydata.RemoveCumulocityProperties(data, true)
		return dst, data, err
	}
}
