package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// WithDeviceGroup device group completion (including root groups and subgroups)
func WithDeviceGroup(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("(name eq '%s') and has(%s)", pattern, "c8y_IsDeviceGroup")
		}))
		return cmd
	}
}

// WithRootDeviceGroup root device group completion
func WithRootDeviceGroup(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, func(pattern string) string {
			return fmt.Sprintf("(name eq '%s') and has(%s) and type eq '%s'", pattern, "c8y_IsDeviceGroup", "c8y_DeviceGroup")
		}))
		return cmd
	}
}
