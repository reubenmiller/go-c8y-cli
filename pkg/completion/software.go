package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// softwareByNameQuery is the inventory query of the software packages with the given name
func softwareByNameQuery(name string) string {
	return fmt.Sprintf("$filter=(name eq '%s') and type eq '%s' $orderby=name,creationTime", name, "c8y_Software")
}

// WithSoftware software (managedObject) completion
func WithSoftware(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, managedObjectCompletion(clientFunc, softwareByNameQuery))
		return cmd
	}
}
