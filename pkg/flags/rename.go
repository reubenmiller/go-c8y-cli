package flags

import (
	"fmt"

	"github.com/spf13/cobra"
)

func MarkRenamed(previous, current string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().MarkDeprecated(previous, fmt.Sprintf("please use --%s instead", current))
		return cmd
	}
}

func MarkDeprecated(cmd *cobra.Command, name string, notice string) {
	if notice == "" {
		notice = "please remove it"
	}
	_ = cmd.Flags().MarkDeprecated(name, notice)
}
