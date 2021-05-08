package flags

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
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
		if value == nil {
			continue
		}
		switch v := value.(type) {
		case iterator.Iterator:
			query.SetVariable(name, v)
			if v.IsBound() {
				totalIterators++
			}
		case map[string]string:
			for key, val := range v {
				query.SetVariable(key, val)
			}
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
				path.SetVariable(name, strings.Join(v, ","))

			case []int:
				path.SetVariable(name, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]"))

			case iterator.Iterator:
				path.SetVariable(name, v)
				if v.IsBound() {
					totalIterators++
				}

			default:
				path.SetVariable(name, fmt.Sprintf("%v", value))
			}
		}
	}

	if err := path.CheckRequired(); err != nil {
		return err
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
		if name != "" {
			switch v := value.(type) {
			case map[string]string:
				for key, val := range v {
					header.Add(key, val)
				}
			default:
				headerValue := fmt.Sprintf("%v", value)
				header.Add(name, headerValue)
			}
		}
	}
	return
}

// WithBody returns a body from given command line arguments
func WithBody(cmd *cobra.Command, body *mapbuilder.MapBuilder, inputIterators *RequestInputIterators, opts ...GetOption) (err error) {
	totalIterators := 0
	iteratorSources := []string{}
	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)
		if err != nil {
			return err
		}

		switch v := value.(type) {
		case iterator.Iterator:
			err = body.Set(name, value)
			if v.IsBound() {
				iteratorSources = append(iteratorSources, name)
				totalIterators++
			}
		case RawString:
			if v != "" {
				body.SetRaw(string(v))
			}
		case string:
			// only set non-empty values by default
			if v != "" {
				err = body.Set(name, value)
			}

		case Template:
			body.AppendTemplate(string(v))
			if body.TemplateIterator == nil {
				body.TemplateIterator = iterator.NewRangeIterator(1, 100000000, 1)
			}

		case TemplateVariables:
			body.SetTemplateVariables(v)

		case DefaultTemplateString:
			// the body will build on this template (it can override it)
			body.PrependTemplate(string(v))

		case RequiredTemplateString:
			// the template will override values in the body
			body.AppendTemplate(string(v))

		case RequiredKeys:
			body.SetRequiredKeys(v...)

		case FilePath:
			if v != "" {
				body.SetFile(string(v))
			}

		case map[string]interface{}:
			if v != nil {
				if name != "" {
					err = body.Set(name, v)
					return err
				} else {
					err = body.MergeMaps(v)
				}
			}
		default:
			if name != "" {
				err = body.Set(name, value)
			}
		}
		if err != nil {
			return err
		}
	}
	if totalIterators > 0 {
		inputIterators.Total += totalIterators
		inputIterators.Body = body

		if len(iteratorSources) > 0 {
			// TODO: Assign values to input template
			body.TemplateIteratorNames = iteratorSources
		}
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
				if !value {
					return "", false, nil
				}
				formattedValue := applyFormatter(format, value)

				if jsonUtilities.IsJSONObject([]byte(formattedValue)) {
					jsonValue := make(map[string]interface{})
					if err := jsonUtilities.ParseJSON(formattedValue, jsonValue); err == nil {
						return dst, jsonValue, err
					}
				}
				return dst, formattedValue, err
			}
			return dst, value, err
		}
		return "", false, nil
	}
}

// WithDefaultBoolValue sets a boolean value regardless if the value has been provided by the flag or not
func WithDefaultBoolValue(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, format := UnpackGetterOptions("", opts...)
		value, err := cmd.Flags().GetBool(src)

		if format != "" {
			if !value {
				return "", false, nil
			}
			formattedValue := applyFormatter(format, value)

			if jsonUtilities.IsJSONObject([]byte(formattedValue)) {
				jsonValue := make(map[string]interface{})
				if err := jsonUtilities.ParseJSON(formattedValue, jsonValue); err == nil {
					return dst, jsonValue, err
				}
			}
			return dst, formattedValue, err
		}
		return dst, value, err
	}
}

// WithOptionalJSONFragment adds fragment if the boolean value is true a boolean value from cli arguments to a query parameter
func WithOptionalJSONFragment(val string, opts ...string) GetOption {
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

		if inputIterators != nil && inputIterators.PipeOptions != nil {
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

// WithCustomStringValue add a custom string value with a custom tranform function
func WithCustomStringValue(transform func([]byte) []byte, targetFunc func() string, opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		src, dst, format := UnpackGetterOptions("%s", opts...)

		if targetFunc != nil {
			if dstName := targetFunc(); dstName != "" {
				dst = dstName
			}
		}

		if inputIterators != nil && inputIterators.PipeOptions != nil {
			if transform != nil {
				inputIterators.PipeOptions.Formatter = transform
			}
			if inputIterators.PipeOptions.Name == src {
				if dst != "" {
					inputIterators.PipeOptions.Property = dst
				}
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
		outputValue := applyFormatter(format, value)
		if transform != nil {
			outputValue = string(transform([]byte(outputValue)))
		}
		return dst, outputValue, err
	}
}

// WithCustomStringSlice adds string  map values from cli arguments
func WithCustomStringSlice(valuesFunc func() ([]string, error), opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		_, dst, format := UnpackGetterOptions("%s", opts...)

		values, err := valuesFunc()
		if err != nil {
			return dst, values, err
		}
		if len(values) == 0 {
			// dont assign the value anywhere
			dst = ""
		}

		outputValues := make(map[string]string)
		for _, v := range values {
			parts := strings.SplitN(v, ":", 2)
			if len(parts) != 2 {
				parts = strings.SplitN(v, "=", 2)
				if len(parts) != 2 {
					continue
				}
			}
			outputValues[strings.TrimSpace(parts[0])] = strings.TrimSpace(applyFormatter(format, parts[1]))
		}

		return dst, outputValues, err
	}
}

// WithOverrideValue adds an options to override a value via cli arguments. Pipeline input is ignored if this value is present
// However if the argument refers to an existing file then the value will be ignored!
func WithOverrideValue(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		src, dst, format := UnpackGetterOptions("%s", opts...)

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			fValue, fErr := cmd.Flags().GetStringSlice(src)
			if fErr != nil || len(fValue) == 0 {
				return dst, value, fErr
			}
			value = fValue[0]
			err = nil
		}
		if value == "" {
			// dont assign the value anywhere
			dst = ""
		}

		if _, err := os.Stat(value); err == nil {
			// ignore input files (they should be piped)
			return "", "", nil
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

// WithStringSliceCSV adds a string slice as comma separated variables from cli arguments
func WithStringSliceCSV(opts ...string) GetOption {
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
		return dst, strings.Join(nonEmptyValues, ","), err
	}
}

// WithIntValue adds a integer (int) value from cli arguments
func WithIntValue(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)

		if inputIterators != nil {
			if inputIterators.PipeOptions.Name == src {
				return WithPipelineIterator(inputIterators.PipeOptions)(cmd, inputIterators)
			}
		}

		value, err := cmd.Flags().GetInt(src)
		return dst, value, err
	}
}

// WithFloatValue adds a float (float32) value from cli arguments
func WithFloatValue(opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("", opts...)

		if inputIterators != nil {
			if inputIterators.PipeOptions.Name == src {
				return WithPipelineIterator(inputIterators.PipeOptions)(cmd, inputIterators)
			}
		}

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

		// mark iterator as unbound, so it will not increment the input iterators
		return dst, iterator.NewRelativeTimeIterator(value, false), err
	}
}

// WithEncodedRelativeTimestamp adds a encoded timestamp (string) value from cli arguments
func WithEncodedRelativeTimestamp(opts ...string) GetOption {
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

		// mark iterator as unbound, so it will not increment the input iterators
		return dst, iterator.NewRelativeTimeIterator(value, true), err
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

// RawString raw string type
type RawString string

// WithDataValueAdvanced adds json or shorthand json parsing with additional option to strip the Cumulocity properties from the input
func WithDataValueAdvanced(stripCumulocityKeys bool, raw bool, opts ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(src) {
			return "", "", nil
		}

		value, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, value, err
		}

		if raw {
			return dst, RawString(resolveContents(value)), nil
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

func WithDataFlagValue() GetOption {
	return WithDataValueAdvanced(true, false, FlagDataName, "")
}

func WithDataValue(opts ...string) GetOption {
	return WithDataValueAdvanced(true, false, opts...)
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
	Name        string              `json:"name"`
	Required    bool                `json:"required"`
	Disabled    bool                `json:"disabled"`
	Property    string              `json:"property"`
	Aliases     []string            `json:"aliases"`
	IsID        bool                `json:"isID"`
	Validator   iterator.Validator  `json:"-"`
	Formatter   func([]byte) []byte `json:"-"`
	InputFilter func([]byte) bool   `json:"-"`
}

// WithPipelineIterator adds pipeline support from cli arguments
func WithPipelineIterator(opts *PipelineOptions) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		iter, err := NewFlagWithPipeIterator(cmd, opts, true)
		if iter == nil {
			return "", iter, err
		}
		return opts.Property, iter, err
	}
}

// NewRequestInputIterators returns input iterations with the pipeline options loaded from the annotations
func NewRequestInputIterators(cmd *cobra.Command) (*RequestInputIterators, error) {
	pipeOpts, err := GetPipeOptionsFromAnnotation(cmd)

	if disableStdin, _ := cmd.Root().PersistentFlags().GetBool(FlagNullInput); disableStdin {
		pipeOpts.Disabled = disableStdin
	}
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
