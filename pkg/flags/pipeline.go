package flags

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/spf13/cobra"
)

var ErrParameterMissing error = errors.New("missing required parameter")

type ParameterError struct {
	Name string

	Err error
}

func (e *ParameterError) Error() string {
	return fmt.Sprintf("Missing required parameter. %s", e.Name)
}

// NewFlagWithPipeIterator creates an iterator from a command argument
// or from the pipeline
// It will automatically try to get the value from a String or a StringSlice flag
func NewFlagWithPipeIterator(cmd *cobra.Command, pipeOpt *PipelineOptions, supportsPipeline bool) (iterator.Iterator, error) {
	if supportsPipeline && !pipeOpt.Disabled {
		sourceProperties := make([]string, 0)
		if len(pipeOpt.Aliases) > 0 {
			sourceProperties = append(sourceProperties, pipeOpt.Aliases...)
		}
		if pipeOpt.Property != "" {
			sourceProperties = append(sourceProperties, pipeOpt.Property)
		}
		iterOpts := &iterator.PipeOptions{
			Properties: sourceProperties,
			Validator:  nil,
			AllowEmpty: !pipeOpt.Required,
			Formatter:  pipeOpt.Formatter,
		}
		lineFilter := FilterJsonLines
		if pipeOpt.InputFilter != nil {
			lineFilter = pipeOpt.InputFilter
		}
		iter, err := iterator.NewJSONPipeIterator(cmd.InOrStdin(), iterOpts, lineFilter)

		if err == iterator.ErrEmptyPipeInput && !pipeOpt.EmptyPipe {
			return iter, err
		}
		if err == nil {
			return iter, nil
		}
	}

	if items, err := GetFlagStringValues(cmd, pipeOpt.Name); err == nil {
		if len(items) == 0 {
			if pipeOpt.Required {
				return nil, &ParameterError{
					Name: pipeOpt.Name,
					Err:  ErrParameterMissing,
				}
			}
		}
		if len(items) > 0 {

			// check if file reference
			if isFile(items[0]) {
				iter, err := iterator.NewFileContentsIterator(items[0])
				if err != nil {
					return nil, err
				}
				return iter, nil
			}

			if pipeOpt.Formatter != nil {
				for i, item := range items {
					items[i] = string(pipeOpt.Formatter([]byte(item)))
				}
			}

			// return array of results
			return iterator.NewSliceIterator(items), nil
		}
	}
	if pipeOpt.Required {
		return nil, fmt.Errorf("no input detected")
	}
	return nil, nil
}

func GetFlagStringValues(cmd *cobra.Command, name string) ([]string, error) {
	items, err := cmd.Flags().GetStringSlice(name)

	if err != nil {
		// fallback to string
		item, strErr := cmd.Flags().GetString(name)

		if strErr != nil {
			return nil, strErr
		}
		if item != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// ErrInvalidIDFormat invalid ID foratm
var ErrInvalidIDFormat = errors.New("invalid id format")

// ValidateID returns an error if the input value does not match an id
func ValidateID(v []byte) (err error) {
	isNotDigit := func(c rune) bool { return c < '0' || c > '9' }
	value := bytes.TrimSpace(v)
	if bytes.IndexFunc(value, isNotDigit) > -1 {
		err = fmt.Errorf("%s. value=%s", ErrInvalidIDFormat, value)
	}
	return
}

func FilterJsonLines(line []byte) bool {
	line = bytes.TrimSpace(line)
	if !bytes.HasPrefix(line, []byte("{")) && !bytes.HasPrefix(line, []byte("[")) {
		return true
	}
	// only allow json objects
	isJSONObject := jsonUtilities.IsJSONObject(line)
	return isJSONObject
}
