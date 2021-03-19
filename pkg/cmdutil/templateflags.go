package cmdutil

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/pathresolver"
	"github.com/spf13/cobra"
)

type TemplatePathResolver struct {
	Path string
}

func (t *TemplatePathResolver) Resolve(name string) (string, error) {
	return matchFilePath(t.Path, name, ".jsonnet", "ignore")
}

func matchFilePath(sourceDir string, pattern string, extension, ignoreDir string) (string, error) {
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
	names, err := pathresolver.ResolvePaths(sourceDir, pattern, extension, ignoreDir)
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

			matches, err := pathresolver.ResolvePaths(templatePath, "*"+toComplete+"*", ".jsonnet", "ignore")
			for i, match := range matches {
				matches[i] = filepath.Base(match)
			}

			if err != nil {
				return []string{"jsonnet"}, cobra.ShellCompDirectiveFilterFileExt
			}
			return matches, cobra.ShellCompDirectiveDefault
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
