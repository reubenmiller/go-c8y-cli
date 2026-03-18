package assert

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	cmdExists "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants/assert/exists"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAssert struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdAssert {
	ccmd := &SubCmdAssert{}

	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Cumulocity tenant assertions",
		Long:  `Assertions for Cumulocity tenants`,
	}

	// Subcommands
	cmd.AddCommand(cmdExists.NewCmdExists(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
