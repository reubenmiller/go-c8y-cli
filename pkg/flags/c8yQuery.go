package flags

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

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

func BuildCumulocityQuery(cmd *cobra.Command, fixedParts []string, orderBy string) func([]byte) []byte {
	return func(b []byte) []byte {
		b = bytes.Replace(b, []byte("$filter="), []byte(""), 1)
		queryParts := fixedParts[:]

		var existingOrderBy []byte
		if idx := bytes.Index(b, []byte("$orderby=")); idx > -1 {
			existingOrderBy = b[idx+9:]
			b = b[:idx]
		} else {
			existingOrderBy = []byte(orderBy)
		}

		if len(b) > 0 {
			queryParts = append(queryParts, string(b))
		}

		if v, err := cmd.Flags().GetString("queryTemplate"); err == nil && v != "" {
			for i := range queryParts {
				queryParts[i] = fmt.Sprintf(v, queryParts[i])
			}
		}

		query := url.QueryEscape(strings.Join(queryParts, " and "))
		query = strings.ReplaceAll(query, "%28", "(")
		query = strings.ReplaceAll(query, "%29", ")")
		query = strings.ReplaceAll(query, "%27", "'")
		orderBy = url.QueryEscape(orderBy)

		outputQuery := []byte(fmt.Sprintf("$filter=%s", query))

		if len(existingOrderBy) > 0 {
			outputQuery = append(outputQuery, []byte(fmt.Sprintf(" $orderby=%s", existingOrderBy))...)
		}
		return outputQuery
		// if len(query) == 0 {
		// 	return []byte(fmt.Sprintf("$filter=%s $orderby=%s", query, orderBy))
		// }
		// return []byte(fmt.Sprintf("$filter=%s $orderby=%s", query, orderBy))
	}
}

// flags.WithCustomStringValue(func(b []byte) []byte {

// 	queryParts := c8yQueryParts[:]
// 	queryParts = append(queryParts, "("+string(b)+")")

// 	if v, err := cmd.Flags().GetString("queryTemplate"); err == nil && v != "" {
// 		for i := range queryParts {
// 			queryParts[i] = fmt.Sprintf(v, queryParts[i])
// 		}
// 	}
// 	query := strings.Join(queryParts, " and ")
// 	return []byte(fmt.Sprintf("$filter=(%s) $orderby=%s", query, orderBy))
// }, func() string {
// 	return "q"
// }, "query")
