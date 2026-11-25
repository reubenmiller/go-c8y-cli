package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WithConfiguration configuration (managedObject) completion
func WithConfiguration(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("(type eq '%s') and (name eq '%s')", "c8y_ConfigurationDump", pattern)
		}))
		return cmd
	}
}
