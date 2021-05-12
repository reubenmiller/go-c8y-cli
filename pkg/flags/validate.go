package flags

import "github.com/spf13/cobra"

// ExactArgsOrExample returns an error if there are not exactly n args or --exapmles is not used
func ExactArgsOrExample(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("examples") {
			return nil
		}
		return cobra.ExactArgs(n)(cmd, args)
	}
}
