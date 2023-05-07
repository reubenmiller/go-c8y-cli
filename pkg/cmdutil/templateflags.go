package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
	"github.com/spf13/cobra"
)

var NamespaceSeparator = "::"

func BuildTemplatePath(namespace, name string) string {
	return fmt.Sprintf("%s%s%s", namespace, NamespaceSeparator, name)
}

type TemplatePathResolver struct {
	Paths []string
}

func (t *TemplatePathResolver) Resolve(name string) (string, error) {
	return matchFilePath(t.Paths, name, []string{".jsonnet"}, "ignore")
}

func matchFilePath(paths []string, pattern string, extensions []string, ignoreDir string) (string, error) {
	// full path
	if stat, err := os.Stat(pattern); err == nil && !stat.IsDir() {
		return pattern, nil
	}

	sourcePattern := ""
	if a, b, ok := strings.Cut(pattern, NamespaceSeparator); ok {
		sourcePattern = a
		pattern = b
	}

	// abort if template path does not exist
	validPaths := []string{}
	for _, sourceDir := range paths {
		sourceName := ""
		sourcePath := sourceDir
		if a, b, ok := strings.Cut(sourceDir, NamespaceSeparator); ok {
			sourceName = a
			sourcePath = b
		}

		if sourcePattern != "" {
			if sourceMatch, _ := filepath.Match(sourceName, sourcePattern); !sourceMatch {
				continue
			}
		}

		if stat, err := os.Stat(sourcePath); err == nil && stat.IsDir() {
			// path exists under template path (return early)
			fullPath := path.Join(sourcePath, pattern)
			if stat, err := os.Stat(fullPath); err == nil && !stat.IsDir() {
				return fullPath, nil
			}

			validPaths = append(validPaths, sourcePath)
		}
	}
	if len(validPaths) == 0 {
		return "", fmt.Errorf("template paths do not exist. %v", paths)
	}

	// try to resolve path in nested
	names, err := pathresolver.ResolvePaths(validPaths, pattern, extensions, ignoreDir)
	if err != nil {
		return "", err
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no matching files found")
	}

	return names[0], nil
}

// WithTemplateFlag add template flag with completion
func (f *Factory) WithTemplateFlag(cmd *cobra.Command) flags.Option {
	return func(cmd *cobra.Command) *cobra.Command {
		// cfg, err := f.Config()
		// if err != nil {
		// 	return nil
		// }
		cmd.Flags().String(flags.FlagDataTemplateName, "", "Body template")
		cmd.Flags().StringArray(flags.FlagDataTemplateVariablesName, []string{}, "Body template variables")

		_ = cmd.RegisterFlagCompletionFunc(flags.FlagDataTemplateName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

			matches, err := f.ResolveTemplates("*"+toComplete+"*", false)

			if err != nil {
				return []string{"jsonnet"}, cobra.ShellCompDirectiveFilterFileExt
			}
			return matches, cobra.ShellCompDirectiveDefault
		})

		_ = cmd.RegisterFlagCompletionFunc(flags.FlagDataTemplateVariablesName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			templateFlag, err := cmd.Flags().GetString(flags.FlagDataTemplateName)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoSpace
			}

			matches, err := f.ResolveTemplates(templateFlag, true)
			// matches, err := pathresolver.ResolvePaths(templatePath, templateFlag, []string{".jsonnet"}, "ignore")

			if err != nil {
				return nil, cobra.ShellCompDirectiveNoSpace
			}

			template := templateFlag
			if len(matches) > 0 {
				template = matches[0]
			}

			if strings.TrimSpace(template) == "" {
				return nil, cobra.ShellCompDirectiveNoSpace
			}

			var scanner *bufio.Scanner
			file, err := os.Open(template)
			if err != nil {
				// assume template is a string
				scanner = bufio.NewScanner(strings.NewReader(template))
			} else {
				scanner = bufio.NewScanner(file)
			}

			pattern := regexp.MustCompile(`var\(['"](.+?)['"](\s*,\s*["']?([^"'\(\)]+)['"]?)?`)
			variableNames := make(map[string]string)
			values := []string{}
			for scanner.Scan() {
				matches := pattern.FindAllStringSubmatch(scanner.Text(), -1)
				for _, match := range matches {
					if len(match) < 2 {
						continue
					}
					if _, ok := variableNames[match[1]]; !ok {
						var value string
						if len(match) >= 4 {
							if match[3] != "" {
								value = fmt.Sprintf("%s=\tdefault: %s", match[1], match[3])
							} else {
								value = fmt.Sprintf("%s=\t(required)", match[1])
							}
						}

						valueKey := match[1]
						variableNames[valueKey] = match[1]
						values = append(values, strings.TrimSpace(value))
					}
				}
			}
			if len(values) == 1 {
				values[0] = strings.Split(values[0], "\t")[0]
			}

			return values, cobra.ShellCompDirectiveNoSpace
		})

		return cmd
	}
}

// WithTemplateValue get the template value using the path resolver controlled by the configuration
func WithTemplateValue(factory *Factory) flags.GetOption {
	return flags.WithTemplateValue(flags.FlagDataTemplateName, NewTemplateResolver(factory))
}

func NewTemplateResolver(factory *Factory) *TemplatePathResolver {
	cfg, err := factory.Config()
	if err != nil {
		return nil
	}
	paths := cfg.GetTemplatePaths()

	for _, ext := range factory.ExtensionManager().List() {
		path := ext.TemplatePath()
		if path != "" {
			paths = append(paths, BuildTemplatePath(ext.Name(), path))
		}
	}

	return &TemplatePathResolver{
		Paths: paths,
	}
}
