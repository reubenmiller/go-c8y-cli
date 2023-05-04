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
				cols := strings.Split(row, ",")
				if len(cols) > 0 {
					options = append(options, cols[0])
				}
			}
			return options, cobra.ShellCompDirectiveNoFileComp

			// TODO: Check if c8y commands can be run directly using the already running instance
			// rootCmd := cmd.Root()
			// cmdArgs := []string{}
			// cmdArgs = append(cmdArgs, externalCommand...)
			// cmdArgs = append(cmdArgs, args...)
			// rootCmd.SetArgs(cmdArgs)

			// stdout := bytes.NewBufferString("")
			// rootCmd.SetOut(stdout)

			// if err := rootCmd.Execute(); err != nil {
			// 	return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			// }

			// results := strings.Split(stdout.String(), "\n")
			// return results, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
