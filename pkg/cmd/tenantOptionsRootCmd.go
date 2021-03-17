package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type TenantOptionsCmd struct {
	*subcommand.SubCommand
}

func NewTenantOptionsRootCmd() *TenantOptionsCmd {
	ccmd := &TenantOptionsCmd{}

	cmd := &cobra.Command{
		Use:   "tenantOptions",
		Short: "Cumulocity tenantOptions",
		Long: `<
REST endpoint to interact with Cumulocity tenantOptions
Options are category-key-value tuples, storing tenant configuration. Some categories of options allow creation of new one, other are limited to predefined set of keys.

Any option of any tenant can be defined as "non-editable" by "management" tenant. Afterwards, any PUT or DELETE requests made on that option by the owner tenant, will result in 403 error (Unauthorized).`,
	}

	// Subcommands
	cmd.AddCommand(NewGetTenantOptionCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewTenantOptionCmd().GetCommand())
	cmd.AddCommand(NewGetTenantOptionCmd().GetCommand())
	cmd.AddCommand(NewDeleteTenantOptionCmd().GetCommand())
	cmd.AddCommand(NewUpdateTenantOptionCmd().GetCommand())
	cmd.AddCommand(NewUpdateTenantOptionBulkCmd().GetCommand())
	cmd.AddCommand(NewGetTenantOptionsForCategoryCmd().GetCommand())
	cmd.AddCommand(NewUpdateTenantOptionEditableCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
