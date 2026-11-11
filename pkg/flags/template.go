package flags

import (
	"fmt"
	"strings"

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

// ResolveTemplate returns the contents of the template file, or the value itself if it
// is not a file. An error is returned if the file is in a denied location
func ResolveTemplate(value string, pathResolver Resolver) (string, error) {
	// ignore errors, as we will try to read the contents
	contents, err := pathResolver.Resolve(value)
	if err != nil {
		return value, nil
	}
	return getContents(contents)
}

func WithTemplateValue(src string, pathResolver Resolver) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		if !cmd.Flags().Changed(src) {
			// ignore
			return "", nil, nil
		}

		value, err := cmd.Flags().GetString(src)
		if err != nil || strings.TrimSpace(value) == "" {
			return "", nil, err
		}

		contents, err := ResolveTemplate(value, pathResolver)
		if err != nil {
			return src, nil, fmt.Errorf("flag: %s. %w", src, err)
		}
		return src, Template(contents), nil
	}
}

func WithTemplateVariablesValue(src ...string) GetOption {
	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		sourceName := FlagDataTemplateVariablesName
		if len(src) > 0 {
			sourceName = src[0]
		}
		opt := WithDataValueAdvanced(false, false, sourceName)
		dst, value, err := opt(cmd, inputIterators)

		if err != nil {
			return dst, value, err
		}

		if dst == "" {
			// ignore value
			return dst, value, err
		}

		switch v := value.(type) {
		case map[string]interface{}:
			return sourceName, TemplateVariables(v), err
		}

		return sourceName, nil, fmt.Errorf("unsupported template variable type")
	}
}

// getContents checks whether the given string is a file reference if so it returns the contents, otherwise it returns the
// input value as is
func getContents(content string) (string, error) {
	return resolveContents(content)
}
