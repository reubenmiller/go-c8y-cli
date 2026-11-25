package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WithFirmware firmware (managedObject) completion
func WithFirmware(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("$filter=(name eq '%s') and type eq '%s' $orderby=name,creationTime", pattern, "c8y_Firmware")
		}))
		return cmd
	}
}
