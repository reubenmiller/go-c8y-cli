package currenttenant

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/currenttenant/get"
	cmdListApplications "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/currenttenant/listapplications"
	cmdVersion "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/currenttenant/version"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
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
	cmd.AddCommand(cmdVersion.NewVersionCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
