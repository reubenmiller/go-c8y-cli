package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WithSmartGroup smart group completion
func WithSmartGroup(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("(name eq '%s') and type eq '%s'", pattern, "c8y_DynamicGroup")
		}))
		return cmd
	}
}
