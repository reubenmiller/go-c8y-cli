package flags

import (
	"fmt"
	"io"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
)

// QueryTemplate is an iterator that can be used to construct HTTP request queries
type QueryTemplate struct {
	templateVariables map[string]interface{}
}

// NewQueryTemplate returns a new query iterator template
func NewQueryTemplate() *QueryTemplate {
	variables := make(map[string]interface{})
	return &QueryTemplate{
		templateVariables: variables,
	}
}

// SetVariable sets a give path variable which will be evalulated when fetching the next value
func (b *QueryTemplate) SetVariable(name string, value interface{}) {
	if name != "" {
		b.templateVariables[name] = value
	}
}

// HasVariable returns true if the variable has already been defined
func (b *QueryTemplate) HasVariable(name string) bool {
	_, ok := b.templateVariables[name]
	return ok
}

// GetQueryUnescape returns the unescaped query. User can choose whether iterators are evalulated or not
func (b *QueryTemplate) GetQueryUnescape(ignoreIterators bool) (string, error) {
	q, err := b.Execute(ignoreIterators)
	if err != nil {
		return "", err
	}
	return url.QueryUnescape(q.Encode())
}

// Execute evalulates the variables and returns a query parameters which can be used for rest requests
func (b *QueryTemplate) Execute(ignoreIterators bool) (query url.Values, err error) {
	query = url.Values{}
	if b.templateVariables == nil {
		return query, io.EOF
	}

	for key, value := range b.templateVariables {

		var currentValue string
		switch v := value.(type) {
		case iterator.Iterator:
			// Unbound iterators are always evaluated!
			if !v.IsBound() || (v.IsBound() && !ignoreIterators) {
				bValue, _, err := v.GetNext()
				if err != nil {
					return query, err
				}
				currentValue = string(bValue)
			}

		case []string:
			// Add each item individually. Don't handle it like the other types
			for _, ivalue := range v {
				if ivalue != "" {
					query.Add(key, fmt.Sprintf("%s", ivalue))
				}
			}

		case string:
			currentValue = v

		default:
			currentValue = fmt.Sprintf("%v", v)
		}
		if currentValue != "" {
			query.Add(key, currentValue)
		}
	}

	return
}

// GetNext returns the next template path
func (b *QueryTemplate) GetNext() ([]byte, interface{}, error) {
	// output, err := b.GetQueryUnescape(false)
	q, err := b.Execute(false)
	if err != nil {
		return nil, nil, err
	}

	output, err := url.QueryUnescape(q.Encode())
	if err != nil {
		return nil, nil, err
	}

	return []byte(output), q, err
}

// IsBound return true if the iterator is bound
func (b *QueryTemplate) IsBound() bool {

	isbound := true
	for _, value := range b.templateVariables {
		switch v := value.(type) {
		case iterator.Iterator:
			if !v.IsBound() {
				isbound = false
				break
			}
		}
	}
	return isbound
}
