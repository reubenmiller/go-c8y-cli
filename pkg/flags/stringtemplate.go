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
	output, _, _ := b.Execute(true)
	return output
}

// Execute replaces all of the path parameters in a given URI with the provided values
// Example:
// 	"alarm/alarms/{id}" => "alarm/alarms/1234" if given a parameter map of {"id": "1234"}
func (b *StringTemplate) Execute(ignoreIterators bool, template ...string) (output string, inputTemplate string, err error) {
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
			if !ignoreIterators {
				nextValue, _, err := v.GetNext()
				if err != nil {
					return "", "", err
				}
				currentValue = string(nextValue)
			}
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
	output, input, err := b.Execute(false)
	return []byte(output), []byte(input), err
}

func replaceVariable(tmpl string, name string, value string) string {
	output := tmpl
	if strings.Count(tmpl, "{") == 1 && strings.Count(tmpl, "}") == 1 {
		// Don't assume the variable name matches the given name.
		// But if there is only one template variable, then it is safe to assume it is the correct one
		output = tmpl[0:strings.Index(tmpl, "{")] + "%s" + tmpl[strings.Index(tmpl, "}")+1:]
	} else {
		// Only substitute an explicitly the pipeline variable name
		output = strings.ReplaceAll(tmpl, "{"+name+"}", "%s")
	}
	return output
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
