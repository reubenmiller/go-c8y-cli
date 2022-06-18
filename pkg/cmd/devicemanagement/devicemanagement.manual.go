package devicemanagement

import (
	cmdCertificates "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement/certificates"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdDeviceManagement struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdDeviceManagement {
	ccmd := &SubCmdDeviceManagement{}

	cmd := &cobra.Command{
		Use:   "devicemanagement",
		Short: "Cumulocity Device management",
		Long:  ``,
	}

	// Subcommands
	cmd.AddCommand(cmdCertificates.NewSubCommand(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
