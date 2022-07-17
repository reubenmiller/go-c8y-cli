package certificates

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement/certificates/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement/certificates/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement/certificates/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement/certificates/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement/certificates/update"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCertificates struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCertificates {
	ccmd := &SubCmdCertificates{}

	cmd := &cobra.Command{
		Use:   "certificates",
		Short: "Device Certificate management",
		Long:  `Manage the trusted certificates which are used by devices.`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
