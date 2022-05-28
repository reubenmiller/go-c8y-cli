package completion

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Option adds flags to a given command
type Option func(*cobra.Command) *cobra.Command

// WithOptions applies given options to the command
func WithOptions(cmd *cobra.Command, opts ...Option) *cobra.Command {
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

// WithValidateSet adds a completion function with the given values
func WithValidateSet(flagName string, values ...string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}

func WithCustomValidateSet(flagName string, customFunc func() []string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return customFunc(), cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}

// WithLazyRequired marks a flag as required but does not enforce it.
func WithLazyRequired(flagName string, values ...string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.Flags().SetAnnotation(flagName, cobra.BashCompOneRequiredFlag, []string{"false"})
		return cmd
	}
}

// MarkLocalFlag marks a flag as local flag so it get prioritized in the completions
func MarkLocalFlag(exclude ...string) Option {
	excludeLookup := map[string]bool{}
	for _, name := range exclude {
		excludeLookup[name] = true
	}
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if _, ok := excludeLookup[f.Name]; !ok {
				_ = cmd.Flags().SetAnnotation(f.Name, cobra.BashCompOneRequiredFlag, []string{"false"})
			}
		})
		return cmd
	}
}

func MatchString(pattern, name string) bool {
	match, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return match
}
