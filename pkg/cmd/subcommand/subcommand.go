package subcommand

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/spf13/cobra"
)

// SubCommand base command containing common command functions
type SubCommand struct {
	cmd *cobra.Command
}

// GetCommand get the cobra command
func (c *SubCommand) GetCommand() *cobra.Command {
	// mark local flags so the appear in completions before global flags
	completion.WithOptions(
		c.cmd,
		completion.MarkLocalFlag(),
	)
	return c.cmd
}

// SubCommand create new sub command
func NewSubCommand(cmd *cobra.Command) *SubCommand {
	return &SubCommand{cmd: cmd}
}
