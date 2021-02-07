package flags

import (
	"io"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/pkg/assert"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
)

func Test_StringTemplateString(t *testing.T) {
	tmpl := NewStringTemplate("/inventory/managedObjects/{id}")
	tmpl.SetVariable("id", "12345")
	out, _, err := tmpl.Execute()
	assert.OK(t, err)
	assert.True(t, out == "/inventory/managedObjects/12345")
}

func Test_StringTemplateInteger(t *testing.T) {
	tmpl := NewStringTemplate("/inventory/managedObjects/{id}")
	tmpl.SetVariable("id", 12345)
	out, _, err := tmpl.Execute()
	assert.OK(t, err)
	assert.True(t, out == "/inventory/managedObjects/12345")
}

func Test_StringTemplateMissingVariable(t *testing.T) {
	tmpl := NewStringTemplate("/inventory/managedObjects/{id}")
	tmpl.SetVariable("name", 12345)
	out, _, err := tmpl.Execute()
	assert.OK(t, err)
	assert.True(t, out == "/inventory/managedObjects/{id}")
}

func Test_StringTemplateMultipleVariable(t *testing.T) {
	tmpl := NewStringTemplate("/inventory/managedObjects/{id}/something/{name}")
	tmpl.SetVariable("id", 12345)
	tmpl.SetVariable("name", "example_name")
	out, _, err := tmpl.Execute()
	assert.OK(t, err)
	assert.True(t, out == "/inventory/managedObjects/12345/something/example_name")
}

func Test_StringTemplateWithIterators(t *testing.T) {
	tmpl := NewStringTemplate("/inventory/managedObjects/{id}/something/{name}")
	tmpl.SetVariable("id", "12345")
	tmpl.SetVariable("name", iterator.NewRepeatIterator("mydevice", 1))
	out, _, err := tmpl.Execute()
	assert.OK(t, err)
	assert.True(t, out == "/inventory/managedObjects/12345/something/mydevice")

	// calling a second time should not be possilbe (as the iterator only has 1 value)
	out, _, err = tmpl.Execute()
	assert.ErrorType(t, err, io.EOF)
}
