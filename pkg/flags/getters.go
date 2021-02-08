package flags

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/pkg/timestamp"
	"github.com/spf13/cobra"
)

// GetOption gets the value from a flag and returns the value which can be set accordingly
type GetOption func(cmd *cobra.Command, inputIterators *RequestInputIterators) (name string, value interface{}, err error)

// WithQueryParameters returns a query parameter values given from command line arguments
func WithQueryParameters(cmd *cobra.Command, query *QueryTemplate, inputIterators *RequestInputIterators, opts ...GetOption) (err error) {
	totalIterators := 0
	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)
		if err != nil {
			return err
		}
		if name == "" {
			continue
		}
		switch v := value.(type) {
		case iterator.Iterator:
			query.SetVariable(name, v)
			totalIterators++
		default:
			query.SetVariable(name, v)
		}
	}
	if totalIterators > 0 {
		inputIterators.Total += totalIterators
		inputIterators.Query = query
	}
	return
}

// WithPathParameters returns a path parameter values given from command line arguments
func WithPathParameters(cmd *cobra.Command, path *StringTemplate, inputIterators *RequestInputIterators, opts ...GetOption) (err error) {
	totalIterators := 0
	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)
		if err != nil {
			return err
		}
		if name != "" {
			switch v := value.(type) {
			case []string:
				path.SetVariable(name, fmt.Sprintf("%s", strings.Join(v, ",")))

			case []int:
				path.SetVariable(name, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]"))

			case iterator.Iterator:
				path.SetVariable(name, v)
				totalIterators++

			default:
				path.SetVariable(name, fmt.Sprintf("%v", value))
			}
		}
	}
	if totalIterators > 0 {
		inputIterators.Total += totalIterators
		inputIterators.Path = path
	}
	return
}

// WithHeaders sets header values from command line arguments
func WithHeaders(cmd *cobra.Command, header http.Header, inputIterators *RequestInputIterators, opts ...GetOption) (err error) {
	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)
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
func WithBody(cmd *cobra.Command, body *mapbuilder.MapBuilder, inputIterators *RequestInputIterators, opts ...GetOption) (err error) {
	totalIterators := 0
	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)
		if err != nil {
			return err
		}

		switch v := value.(type) {
		case iterator.Iterator:
			err = body.Set(name, value)
			totalIterators++
		case string:
			// only set non-empty values by default
			if v != "" {
				err = body.Set(name, value)
			}

		case Template:
			body.SetApplyTemplateOnMarshalPreference(true)
			body.SetTemplate(string(v))
			if body.TemplateIterator == nil {
				body.TemplateIterator = iterator.NewRangeIterator(1, 100000000, 1)
			}

		case TemplateVariables:
			body.SetTemplateVariables(v)

		case DefaultTemplateString:
			// the body will build on this template (it can override it)
			err = body.MergeJsonnet(string(v), true)

		case RequiredTemplateString:
			// the template will override values in the body
			err = body.MergeJsonnet(string(v), false)

		case RequiredKeys:
			body.SetRequiredKeys(v...)

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
	if totalIterators > 0 {
		inputIterators.Total += totalIterators
		inputIterators.Body = body
	}
	return nil
}

// WithBoolValue adds a boolean value from cli arguments to a query parameter
func WithBoolValue(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
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
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		src, dst, format := UnpackGetterOptions("%s", opts...)

		if inputIterators != nil {
			if inputIterators.PipeOptions.Name == src {
				return WithPipelineIterator(inputIterators.PipeOptions)(cmd, inputIterators)
			}
		}

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}
		if value == "" {
			// dont assign the value anywhere
			dst = ""
		}
		return dst, applyFormatter(format, value), err
	}
}

// WithStringDefaultValue adds a string value from cli arguments
func WithStringDefaultValue(defaultValue string, opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

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
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

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
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)
		value, err := cmd.Flags().GetInt(src)
		return dst, value, err
	}
}

// WithFloatValue adds a float (float32) value from cli arguments
func WithFloatValue(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)
		value, err := cmd.Flags().GetFloat32(src)
		return dst, value, err
	}
}

// WithRelativeTimestamp adds a timestamp (string) value from cli arguments
func WithRelativeTimestamp(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
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
		if options[2] != "" {
			formatter = options[2]
		}
	}
	return
}

// FilePath is a string representation of a file path
type FilePath string

// WithFilePath adds a file path from cli arguments
func WithFilePath(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := UnpackGetterOptions("%s", opts...)

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}

		return dst, FilePath(value), err
	}
}

// WithDataValueAdvanced adds json or shorthand json parsing with additional option to strip the Cumulocity properties from the input
func WithDataValueAdvanced(stripCumulocityKeys bool, opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(src) {
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

		if stripCumulocityKeys {
			c8ydata.RemoveCumulocityProperties(data, true)
		}
		return dst, data, err
	}
}

func WithDataValue(opts ...string) GetOption {
	return WithDataValueAdvanced(true, opts...)
}

type DefaultTemplateString string
type RequiredTemplateString string

func WithTemplateString(value string, applyLast bool) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		if applyLast {
			return "", RequiredTemplateString(value), nil
		}
		return "", DefaultTemplateString(value), nil
	}
}

func WithDefaultTemplateString(value string) GetOption {
	return WithTemplateString(value, false)
}

func WithRequiredTemplateString(value string) GetOption {
	return WithTemplateString(value, true)
}

type RequiredKeys []string

func WithRequiredProperties(values ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		return "", RequiredKeys(values), nil

	}
}

type PipelineOptions struct {
	Name     string   `json:"name"`
	Required bool     `json:"required"`
	Property string   `json:"property"`
	Aliases  []string `json:"aliases"`
}

// WithPipelineIterator adds pipeline support from cli arguments
func WithPipelineIterator(opts *PipelineOptions) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		iter, err := NewFlagWithPipeIterator(cmd, opts, true)
		return opts.Property, iter, err
	}
}

// NewRequestInputIterators returns input iterations with the pipeline options loaded from the annotations
func NewRequestInputIterators(cmd *cobra.Command) (*RequestInputIterators, error) {
	pipeOpts, err := GetPipeOptionsFromAnnotation(cmd)
	inputIter := &RequestInputIterators{
		PipeOptions: pipeOpts,
	}
	return inputIter, err
}

// RequestInputIterators contains all request input iterators
type RequestInputIterators struct {
	Total       int
	Path        *StringTemplate
	Body        *mapbuilder.MapBuilder
	Query       *QueryTemplate
	PipeOptions *PipelineOptions
}
