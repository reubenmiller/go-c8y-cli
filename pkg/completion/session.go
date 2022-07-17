package completion

import (
	"path/filepath"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
	"github.com/spf13/cobra"
)

// WithSessionFile session file completion
func WithSessionFile(flagName string, extensions []string, pathFunc func() string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			sourcePath := pathFunc()

			matches, err := pathresolver.ResolvePaths([]string{sourcePath}, "*"+toComplete+"*", extensions, "ignore")
			for i, match := range matches {
				matches[i] = filepath.Base(match)
			}

			if err != nil {
				return []string{"json"}, cobra.ShellCompDirectiveFilterFileExt
			}
			return matches, cobra.ShellCompDirectiveNoSpace
		})
		return cmd
	}
}
