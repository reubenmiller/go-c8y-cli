package flags

import (
	"fmt"
	"io"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
)

// StringTemplate is a template which can input parameters which will be ev
type StringTemplate struct {
	template          string
	templateVariables map[string]interface{}
}

// NewStringTemplate returns a new string template
func NewStringTemplate(template string) *StringTemplate {
	variables := make(map[string]interface{})
	return &StringTemplate{
		template:          template,
		templateVariables: variables,
	}
}

// SetVariable sets a give path variable which will be evalulated when fetching the next value
func (b *StringTemplate) SetVariable(name string, value interface{}) {
	b.templateVariables[name] = value
}

// GetTemplate return the string template
func (b *StringTemplate) GetTemplate() string {
	return b.template
}

// Execute replaces all of the path parameters in a given URI with the provided values
// Example:
// 	"alarm/alarms/{id}" => "alarm/alarms/1234" if given a parameter map of {"id": "1234"}
func (b *StringTemplate) Execute(template ...string) (output string, inputTemplate string, err error) {
	if b.templateVariables == nil {
		return "", "", io.EOF
	}

	inputTemplate = b.template
	if len(template) > 0 {
		inputTemplate = template[0]
	}
	output = inputTemplate

	for key, value := range b.templateVariables {

		var currentValue string
		switch v := value.(type) {
		case iterator.Iterator:
			nextValue, _, err := v.GetNext()
			if err != nil {
				return "", "", err
			}
			currentValue = string(nextValue)
		case string:
			currentValue = v
		default:
			currentValue = fmt.Sprintf("%v", v)
		}
		output = strings.ReplaceAll(output, "{"+key+"}", currentValue)
	}

	return
}

// GetNext returns the next template path
func (b *StringTemplate) GetNext() ([]byte, interface{}, error) {

	output, input, err := b.Execute()
	return []byte(output), []byte(input), err
}
