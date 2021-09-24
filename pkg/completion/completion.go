package completion

import (
	"github.com/spf13/cobra"
)

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
