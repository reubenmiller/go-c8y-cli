package flags

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/spf13/cobra"
)

// WithC8YQueryOptions applies given options to a URL Query parameters
func WithC8YQueryOptions(cmd *cobra.Command, inputIterators *RequestInputIterators, opts ...GetOption) ([]string, error) {
	queryParts := make([]string, 0)
	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)

		if err != nil {
			return queryParts, err
		}

		// use the name as an indicator of an empty value
		if name == "" {
			continue
		}

		switch v := value.(type) {
		case string:
			if v != "" {
				queryParts = append(queryParts, v)
			}
		case iterator.Iterator:
			nextvalue, _, err := v.GetNext()
			if err != nil && err != io.EOF {
				return queryParts, err
			}
			if len(nextvalue) > 0 {
				queryParts = append(queryParts, string(nextvalue))
			}
		}
		if err != nil {
			return queryParts, err
		}
	}
	return queryParts, nil
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

		if tmpl, err := cmd.Flags().GetString("queryTemplate"); err == nil && tmpl != "" {
			if len(b) > 0 {
				queryParts = append(queryParts, fmt.Sprintf(tmpl, b))
			} else {
				queryParts = append(queryParts, tmpl)
			}
		} else {
			if len(b) > 0 {
				// Template is not defined so use value as is
				queryParts = append(queryParts, string(b))
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
	}
}
