package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
	"github.com/spf13/cobra"
)

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

	// abort if template path does not exist
	validPaths := []string{}
	for _, sourceDir := range paths {

		if stat, err := os.Stat(sourceDir); err == nil && stat.IsDir() {
			// path exists under template path (return early)
			fullPath := path.Join(sourceDir, pattern)
			if stat, err := os.Stat(fullPath); err == nil && !stat.IsDir() {
				return fullPath, nil
			}

			validPaths = append(validPaths, sourceDir)
		}
	}
	if len(validPaths) == 0 {
		return "", fmt.Errorf("Template paths do not exist. %v", paths)
	}

	// try to resolve path in nested
	names, err := pathresolver.ResolvePaths(paths, pattern, extensions, ignoreDir)
	if err != nil {
		return "", err
	}

	if len(names) == 0 {
		return "", fmt.Errorf("No matching files found")
	}

	return names[0], nil
}

// WithTemplateFlag add template flag with completion
func (f *Factory) WithTemplateFlag(cmd *cobra.Command) flags.Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cfg, err := f.Config()
		if err != nil {
			return nil
		}
		cmd.Flags().String(flags.FlagDataTemplateName, "", "Body template")
		cmd.Flags().String(flags.FlagDataTemplateVariablesName, "", "Body template variables")

		_ = cmd.RegisterFlagCompletionFunc(flags.FlagDataTemplateName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			templatePath := cfg.GetTemplatePaths()

			matches, err := pathresolver.ResolvePaths(templatePath, "*"+toComplete+"*", []string{".jsonnet"}, "ignore")
			for i, match := range matches {
				matches[i] = filepath.Base(match)
			}

			if err != nil {
				return []string{"jsonnet"}, cobra.ShellCompDirectiveFilterFileExt
			}
			return matches, cobra.ShellCompDirectiveDefault
		})

		_ = cmd.RegisterFlagCompletionFunc(flags.FlagDataTemplateVariablesName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			templatePath := cfg.GetTemplatePaths()
			templateFlag, err := cmd.Flags().GetString(flags.FlagDataTemplateName)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoSpace
			}
			matches, err := pathresolver.ResolvePaths(templatePath, templateFlag, []string{".jsonnet"}, "ignore")

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
func WithTemplateValue(cfg *config.Config) flags.GetOption {
	resolve := &TemplatePathResolver{
		Paths: cfg.GetTemplatePaths(),
	}
	return flags.WithTemplateValue(flags.FlagDataTemplateName, resolve)
}
