package flags

import (
	"bytes"
	"fmt"
	"os"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/spf13/cobra"
)

// NewFlagWithPipeIterator creates an iterator from a command argument
// or from the pipeline
// It will automatically try to get the value from a String or a StringSlice flag
func NewFlagWithPipeIterator(cmd *cobra.Command, pipeOpt *PipelineOptions, supportsPipeline bool) (iterator.Iterator, error) {
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
	} else if supportsPipeline {
		sourceProperties := make([]string, 0)
		if pipeOpt.Property != "" {
			sourceProperties = append(sourceProperties, pipeOpt.Property)
		}
		if len(pipeOpt.Aliases) > 0 {
			sourceProperties = append(sourceProperties, pipeOpt.Aliases...)
		}
		iterOpts := &iterator.PipeOptions{
			Properties: sourceProperties,
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
		if err != nil {
			if pipeOpt.Required {
				return iter, err
			}
			return iter, nil
		}
		return iter, nil
	}
	if pipeOpt.Required {
		return nil, fmt.Errorf("no input detected")
	}
	return nil, nil
}
