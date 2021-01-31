package flags

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

// QueryOption adds values to the query parameter based on the command arguments set
type QueryOption func(*cobra.Command, url.Values) error

// WithQueryOptions applies given options to a URL Query parameters
func WithQueryOptions(cmd *cobra.Command, query url.Values, opts ...QueryOption) error {
	for _, opt := range opts {
		if err := opt(cmd, query); err != nil {
			return err
		}
	}
	return nil
}

var ErrFlagDoesNotExist = errors.New("Flag does not exist")
var ErrInvalidArgumentCount = errors.New("invalid number of arguments")

// WithQueryBool adds a boolean value from cli arguments to a query parameter
func WithQueryBool(name ...string) QueryOption {
	return func(cmd *cobra.Command, query url.Values) error {
		src, dst := getNameMapping(name...)
		if cmd.Flags().Changed(src) {
			if v, err := cmd.Flags().GetBool(src); err == nil {
				query.Add(dst, fmt.Sprintf("%v", v))
			} else {
				return ErrFlagDoesNotExist
			}
		}
		return nil
	}
}

// WithQueryString adds a string value from cli arguments to a query parameter
func WithQueryString(name ...string) QueryOption {
	return func(cmd *cobra.Command, query url.Values) error {
		src, dst := getNameMapping(name...)
		if cmd.Flags().Changed(src) {
			if v, err := cmd.Flags().GetString(src); err == nil {
				query.Add(dst, fmt.Sprintf("%v", v))
			} else {
				return ErrFlagDoesNotExist
			}
		}
		return nil
	}
}

func WithCurrentPage() QueryOption {
	return WithQueryString("currentPage")
}

func WithPageSize() QueryOption {
	return WithQueryString("pageSize")
}

func WithTotalPages() QueryOption {
	return WithQueryString("withTotalPages")
}

func getNameMapping(name ...string) (string, string) {
	var src string
	var dst string
	if len(name) == 1 {
		src = name[0]
		dst = name[0]
	} else if len(name) >= 2 {
		src = name[0]
		dst = name[1]
	}
	return src, dst
}
