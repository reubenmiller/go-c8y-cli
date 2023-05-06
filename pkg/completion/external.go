package completion

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extexec"
	"github.com/spf13/cobra"
)

// WithExternalCompletion completion by executing an external command or another c8y command
func WithExternalCompletion(flagName string, externalCommand []string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

			log.Printf("Completing external flag. args=%v, completion_cmd=%v, toComplete=%s\n", os.Args, externalCommand, toComplete)
			output, err := extexec.ExecuteExternalCommand(toComplete, externalCommand)

			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					log.Printf("stderr: %s", exiterr.Stderr)
				} else {
					log.Printf("output: %s, %s", output, err)
				}

				return []string{err.Error()}, cobra.ShellCompDirectiveNoFileComp
			}

			log.Printf("Output: %s", output)
			// TODO: Use tsv instead of csv output
			// TODO: Add support for field annotations on the descriptions
			// eg. // rowParts = append(rowParts, fmt.Sprintf("%s: %s", fields[i], col))
			options := []string{}
			for _, row := range strings.Split(string(output), "\n") {
				if len(row) > 0 {
					options = append(options, strings.ReplaceAll(row, ",", "\t"))
				}
			}
			return options, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
