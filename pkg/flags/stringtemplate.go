package flags

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
)

// StringTemplate is a template which can input parameters which will be ev
type StringTemplate struct {
	template          string
	templateVariables map[string]interface{}
	allowEmptyValues  bool
}

// NewStringTemplate returns a new string template
func NewStringTemplate(template string) *StringTemplate {
	variables := make(map[string]interface{})
	return &StringTemplate{
		template:          template,
		templateVariables: variables,
	}
}

// SetTemplate updates the string template
func (b *StringTemplate) SetTemplate(template string) {
	b.template = template
}

// SetVariable sets a give path variable which will be evalulated when fetching the next value
func (b *StringTemplate) SetVariable(name string, value interface{}) {
	b.templateVariables[name] = value
}

func (b *StringTemplate) SetAllowEmptyValues(value bool) {
	b.allowEmptyValues = value
}

// GetTemplate return the string template
func (b *StringTemplate) GetTemplate() string {
	output, _, _ := b.Execute(true)
	return output
}

// CheckRequired check if all required variables are provided
func (b *StringTemplate) CheckRequired() error {
	pattern := regexp.MustCompile(`\{(\w+)\}`)

	missingNames := []string{}
	for _, names := range pattern.FindAllStringSubmatch(b.template, -1) {
		if _, ok := b.templateVariables[names[1]]; !ok {
			missingNames = append(missingNames, names[1])
		}
	}

	if len(missingNames) == 0 {
		return nil
	}

	return fmt.Errorf("missing required parameters. %v", missingNames)
}

// Execute replaces all of the path parameters in a given URI with the provided values
// Example:
// 	"alarm/alarms/{id}" => "alarm/alarms/1234" if given a parameter map of {"id": "1234"}
func (b *StringTemplate) Execute(ignoreIterators bool, template ...string) (output string, input interface{}, err error) {
	if b.templateVariables == nil {
		return "", "", io.EOF
	}

	output = b.template
	if len(template) > 0 {
		output = template[0]
	}
	input = output

	for key, value := range b.templateVariables {

		var currentValue string
		switch v := value.(type) {
		case iterator.Iterator:
			// Always evaluate unbound iterators
			if !v.IsBound() || (v.IsBound() && !ignoreIterators) {
				nextValue, inputValue, err := v.GetNext()
				if err != nil {
					return "", "", err
				}

				input = inputValue
				currentValue = string(nextValue)
			}
		case string:
			currentValue = v
		default:
			currentValue = fmt.Sprintf("%v", v)
		}
		if b.allowEmptyValues || currentValue != "" {
			output = strings.ReplaceAll(output, "{"+key+"}", currentValue)
		}
	}

	return
}

// GetNext returns the next template path
func (b *StringTemplate) GetNext() ([]byte, interface{}, error) {
	output, input, err := b.Execute(false)
	return []byte(output), input, err
}

// IsBound return true if the iterator is bound
func (b *StringTemplate) IsBound() bool {

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
