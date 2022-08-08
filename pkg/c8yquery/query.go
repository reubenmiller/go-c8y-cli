package c8yquery

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/spf13/cobra"
)

type QueryPart struct {
	Name  string
	Value interface{}
}

type CumulocityQueryIterator struct {
	Filter     []QueryPart
	FilterJoin string
	OrderBy    []string
}

func NewCumulocityQueryIterator() *CumulocityQueryIterator {
	return &CumulocityQueryIterator{
		FilterJoin: " and ",
	}
}

func (i *CumulocityQueryIterator) GetNext() (line []byte, input interface{}, err error) {
	queryParts := make([]string, 0)
	for _, part := range i.Filter {

		switch value := part.Value.(type) {
		case string:
			if value != "" {
				queryParts = append(queryParts, value)
			}
		case iterator.Iterator:
			line, _, err := value.GetNext()
			if err != nil {
				return nil, nil, err
			}
			if len(line) > 0 {
				queryParts = append(queryParts, string(line))
			}
		}
	}

	filter := url.QueryEscape(strings.Join(queryParts, i.FilterJoin))

	orderBy := ""
	if len(i.OrderBy) > 0 {
		orderBy = "+$orderby=" + url.QueryEscape(strings.Join(i.OrderBy, " "))
	}

	return []byte(fmt.Sprintf("$filter=%s%s", filter, orderBy)), "", nil
}

func (i *CumulocityQueryIterator) IsBound() bool {
	for _, part := range i.Filter {
		if v, ok := part.Value.(iterator.Iterator); ok {
			if v.IsBound() {
				return true
			}
		}
	}
	return false
}

func (i *CumulocityQueryIterator) AddFilterPart(name string, value interface{}) {
	i.Filter = append(i.Filter, QueryPart{name, value})
}

func (i *CumulocityQueryIterator) AddOrderPart(order string) {
	i.OrderBy = append(i.OrderBy, order)
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
