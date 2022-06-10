package subcommand

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

// SubCommand base command containing common command functions
type SubCommand struct {
	cmd           *cobra.Command
	requiredFlags []string
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

func (c *SubCommand) SetRequiredFlags(v ...string) *SubCommand {
	c.requiredFlags = v
	return c
}

func (c *SubCommand) CheckRequiredFlags() error {
	for _, name := range c.requiredFlags {
		if !c.cmd.Flags().Changed(name) {
			return &flags.ParameterError{
				Name: name,
				Err:  flags.ErrParameterMissing,
			}
		}
	}
	return nil
}

// SubCommand create new sub command
func NewSubCommand(cmd *cobra.Command) *SubCommand {
	subcmd := &SubCommand{cmd: cmd}

	// Wrap RunE so that the custom required flags can be called each time
	// This is required due to the hack for the tab completion which prioritizes required
	// parameter before other local parameters. Otherwise the user is not presented with the list of options.
	if cmd.RunE != nil {
		origRunE := cmd.RunE
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			if err := subcmd.CheckRequiredFlags(); err != nil {
				return err
			}
			return origRunE(cmd, args)
		}
	}

	return subcmd
}
