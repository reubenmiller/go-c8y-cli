package completion

import (
	"context"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

func WithDisabledDryRunContext(c *c8y.Client) context.Context {
	return c.Context.CommonOptions(c8y.CommonOptions{
		DryRun: false,
	})
}

// GetFlagStringValues get string slice from either a string slice or string flag
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
