package flags

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

type Template string
type TemplateVariables map[string]interface{}

type Resolver interface {
	Resolve(string) (string, error)
}

func WithTemplateOptions(templateName string, variablesName string, pathResolver Resolver) []GetOption {
	return []GetOption{
		WithTemplateValue(templateName, pathResolver),
		WithTemplateVariablesValue(variablesName),
	}
}

func WithTemplateValue(src string, pathResolver Resolver) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {
		if !cmd.Flags().Changed(src) {
			// ignore
			return "", nil, nil
		}

		value, err := cmd.Flags().GetString(FlagDataTemplateVariablesName)
		if err != nil {
			return "", nil, err
		}

		fullFilePath, err := pathResolver.Resolve(value)
		if err != nil {
			return "", nil, err
		}

		content := getContents(fullFilePath)
		return src, content, nil
	}
}

func WithTemplateVariablesValue(src string) GetOption {
	return func(cmd *cobra.Command) (string, interface{}, error) {

		opt := WithDataValueAdvanced(false, src)
		dst, value, err := opt(cmd)

		if err != nil {
			return dst, value, err
		}

		if dst == "" {
			// ignore value
			return dst, value, err
		}

		switch v := value.(type) {
		case map[string]interface{}:
			return src, TemplateVariables(v), err
		}

		return src, nil, fmt.Errorf("unsupported template variable type")
	}
}

// getContents checks whether the given string is a file reference if so it returns the contents, otherwise it returns the
// input value as is
func getContents(content string) string {
	if _, err := os.Stat(content); err != nil {
		// not a file
		return content
	}

	fileContent, err := ioutil.ReadFile(content)
	if err != nil {
		return content
	}
	// file contents
	return string(fileContent)
}
