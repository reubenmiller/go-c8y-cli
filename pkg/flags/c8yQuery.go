package flags

import (
	"fmt"

	"github.com/spf13/cobra"
)

// C8YQueryOption adds values to the query parameter based on the command arguments set
type C8YQueryOption func(cmd *cobra.Command) (string, error)

// WithC8YQueryOptions applies given options to a URL Query parameters
func WithC8YQueryOptions(cmd *cobra.Command, opts ...C8YQueryOption) ([]string, error) {
	queryParts := make([]string, 0)
	for _, opt := range opts {
		v, err := opt(cmd)
		if err != nil {
			return queryParts, err
		}
		if v != "" {
			queryParts = append(queryParts, v)
		}
	}
	return queryParts, nil
}

// WithC8YQueryFormat adds support for a format string
func WithC8YQueryFormat(name, format string) C8YQueryOption {
	return func(cmd *cobra.Command) (string, error) {
		if v, err := cmd.Flags().GetString(name); err == nil {
			if v != "" {
				return fmt.Sprintf(format, v), nil
			}
		} else {
			return "", ErrFlagDoesNotExist
		}
		return "", nil
	}
}

// WithC8YQueryBool adds support for a fixed string based on a boolean argument
func WithC8YQueryBool(name, value string) C8YQueryOption {
	return func(cmd *cobra.Command) (string, error) {
		if v, err := cmd.Flags().GetBool(name); err == nil {
			if v {
				return value, nil
			}
		} else {
			return "", ErrFlagDoesNotExist
		}
		return "", nil
	}
}

// WithC8YQueryFixedString returns a fixed string value
func WithC8YQueryFixedString(value string) C8YQueryOption {
	return func(cmd *cobra.Command) (string, error) {
		return value, nil
	}
}
