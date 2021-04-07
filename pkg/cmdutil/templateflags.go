package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/pathresolver"
	"github.com/spf13/cobra"
)

type TemplatePathResolver struct {
	Path string
}

func (t *TemplatePathResolver) Resolve(name string) (string, error) {
	return matchFilePath(t.Path, name, []string{".jsonnet"}, "ignore")
}

func matchFilePath(sourceDir string, pattern string, extensions []string, ignoreDir string) (string, error) {
	// full path
	if _, err := os.Stat(pattern); err == nil {
		return pattern, nil
	}

	// abort if template path does not exist
	if _, err := os.Stat(sourceDir); err != nil {
		return "", fmt.Errorf("Template path does not exist. %s", sourceDir)
	}

	// path exists under template path
	fullPath := path.Join(sourceDir, pattern)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// try to resolve path in nested
	names, err := pathresolver.ResolvePaths(sourceDir, pattern, extensions, ignoreDir)
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
			templatePath := cfg.GetTemplatePath()

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
			templatePath := cfg.GetTemplatePath()
			templateFlag, err := cmd.Flags().GetString(flags.FlagDataTemplateName)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoSpace
			}
			matches, err := pathresolver.ResolvePaths(templatePath, templateFlag, []string{".jsonnet"}, "ignore")

			if err != nil {
				return nil, cobra.ShellCompDirectiveNoSpace
			}

			templateFile := ""
			if len(matches) > 0 {
				templateFile = matches[0]
			}

			if strings.TrimSpace(templateFile) == "" {
				return nil, cobra.ShellCompDirectiveNoSpace
			}

			file, err := os.Open(templateFile)
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveNoSpace
			}

			scanner := bufio.NewScanner(file)
			pattern := regexp.MustCompile(`var\("(.+?)"(\s*,\s*"?([^"\(\)]+)"?)?`)
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
		Path: cfg.GetTemplatePath(),
	}
	return flags.WithTemplateValue(flags.FlagDataTemplateName, resolve)
}
