package bulk

import (
	cmdRegister "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration/bulk/register"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdBulk struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdBulk {
	ccmd := &SubCmdBulk{}

	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Cumulocity bulk device registration",
	}

	// Subcommands
	cmd.AddCommand(cmdRegister.NewRegisterCmd(f).GetCommand())
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
