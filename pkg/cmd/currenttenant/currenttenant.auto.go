package currenttenant

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currenttenant/get"
	cmdGetVersion "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currenttenant/getversion"
	cmdListApplications "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currenttenant/listapplications"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCurrenttenant struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCurrenttenant {
	ccmd := &SubCmdCurrenttenant{}

	cmd := &cobra.Command{
		Use:   "currenttenant",
		Short: "Cumulocity current tenant",
		Long:  `Cumulocity current tenant commands`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdListApplications.NewListApplicationsCmd(f).GetCommand())
	cmd.AddCommand(cmdGetVersion.NewGetVersionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
