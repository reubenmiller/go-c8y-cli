package cmd

import (
	"github.com/spf13/cobra"
)

type TenantOptionsCmd struct {
	*baseCmd
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
	cmd.AddCommand(NewGetTenantOptionCollectionCmd().getCommand())
	cmd.AddCommand(NewNewTenantOptionCmd().getCommand())
	cmd.AddCommand(NewGetTenantOptionCmd().getCommand())
	cmd.AddCommand(NewDeleteTenantOptionCmd().getCommand())
	cmd.AddCommand(NewUpdateTenantOptionCmd().getCommand())
	cmd.AddCommand(NewUpdateTenantOptionBulkCmd().getCommand())
	cmd.AddCommand(NewGetTenantOptionsForCategoryCmd().getCommand())
	cmd.AddCommand(NewUpdateTenantOptionEditableCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
