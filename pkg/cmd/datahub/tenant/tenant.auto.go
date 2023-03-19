package tenant

import (
	cmdGet "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/datahub/tenant/get"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTenant struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTenant {
	ccmd := &SubCmdTenant{}

	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Cumulocity IoT DataHub Tenant information",
		Long:  `Cumulocity IoT DataHub Tenant information`,
	}

	// Subcommands
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
