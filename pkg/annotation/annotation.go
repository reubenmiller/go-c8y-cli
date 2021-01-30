package annotation

import (
	"strings"

	"github.com/spf13/cobra"
)

// FlagValueFromPipeline is an annotation which indicates if the flag supported piped input or not
var FlagValueFromPipeline = "valueFromPipeline"

// HasValueFromPipeline checks if the given flag name supported values from pipeline
// It checks the command for a special annotation
func HasValueFromPipeline(cmd *cobra.Command, name string) bool {
	if cmd.Annotations != nil {
		if pipedArgName, ok := cmd.Annotations[FlagValueFromPipeline]; ok {
			return strings.EqualFold(pipedArgName, name)
		}
	}
	return false
}
