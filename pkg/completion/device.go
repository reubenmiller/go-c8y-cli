package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WithDevice device completion
func WithDevice(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("(name eq '%s') and has(%s)", pattern, "c8y_IsDevice")
		}))
		return cmd
	}
}

// WithAgent agent completion
func WithAgent(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("(name eq '%s') and has(%s)", pattern, "com_cumulocity_model_Agent")
		}))
		return cmd
	}
}
