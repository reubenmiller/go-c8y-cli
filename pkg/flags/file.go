package flags

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/spf13/cobra"
)

// NewFlagFileContents returns iterator which will interate over the lines in a file
func NewFlagFileContents(cmd *cobra.Command, name string) (iterator.Iterator, error) {
	supportsPipeline := HasValueFromPipeline(cmd, name)
	if cmd.Flags().Changed(name) {
		if path, err := cmd.Flags().GetString(name); err == nil && path != "" {

			iter, err := iterator.NewFileContentsIterator(path)

			if err != nil {
				return nil, err
			}
			return iter, nil
		}
	} else if supportsPipeline {
		return iterator.NewPipeIterator()
	}
	return nil, fmt.Errorf("no input detected")
}
