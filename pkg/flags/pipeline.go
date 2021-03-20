package flags

import (
	"bytes"
	"errors"
	"fmt"
	"os"

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
	if supportsPipeline {
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
		}
		iter, err := iterator.NewJSONPipeIterator(cmd.InOrStdin(), iterOpts, func(line []byte) bool {
			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("{")) && !bytes.HasPrefix(line, []byte("[")) {
				return true
			}
			// only allow json objects
			isJSONObject := jsonUtilities.IsJSONObject(line)
			return isJSONObject
		})
		if err == nil {
			return iter, nil
		}
	}

	if cmd.Flags().Changed(pipeOpt.Name) {

		items, err := cmd.Flags().GetStringSlice(pipeOpt.Name)

		if err != nil {
			// fallback to string
			item, err := cmd.Flags().GetString(pipeOpt.Name)

			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
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
			if _, err := os.Stat(items[0]); err == nil {
				iter, err := iterator.NewFileContentsIterator(items[0])
				if err != nil {
					return nil, err
				}
				return iter, nil
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
