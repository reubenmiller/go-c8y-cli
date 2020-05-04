package cmd

import (
	"github.com/spf13/cobra"
)

type tenantOptionsCmd struct {
	*baseCmd
}

func newTenantOptionsRootCmd() *tenantOptionsCmd {
	ccmd := &tenantOptionsCmd{}

	cmd := &cobra.Command{
		Use:   "tenantOptions",
		Short: "Cumulocity tenantOptions",
		Long: `<
REST endpoint to interact with Cumulocity tenantOptions
Options are category-key-value tuples, storing tenant configuration. Some categories of options allow creation of new one, other are limited to predefined set of keys.

Any option of any tenant can be defined as "non-editable" by "management" tenant. Afterwards, any PUT or DELETE requests made on that option by the owner tenant, will result in 403 error (Unauthorized).`,
	}

	// Subcommands
	cmd.AddCommand(newGetTenantOptionCollectionCmd().getCommand())
	cmd.AddCommand(newNewTenantOptionCmd().getCommand())
	cmd.AddCommand(newGetTenantOptionCmd().getCommand())
	cmd.AddCommand(newDeleteTenantOptionCmd().getCommand())
	cmd.AddCommand(newUpdateTenantOptionCmd().getCommand())
	cmd.AddCommand(newUpdateTenantOptionBulkCmd().getCommand())
	cmd.AddCommand(newGetTenantOptionsForCategoryCmd().getCommand())
	cmd.AddCommand(newUpdateTenantOptionEditableCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
