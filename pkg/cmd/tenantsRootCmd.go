package cmd

import (
	"github.com/spf13/cobra"
)

type tenantsCmd struct {
	*baseCmd
}

func newTenantsRootCmd() *tenantsCmd {
	ccmd := &tenantsCmd{}

	cmd := &cobra.Command{
		Use:   "tenants",
		Short: "Cumulocity tenant",
		Long:  `REST endpoint to interact with Cumulocity tenants`,
	}

	// Subcommands
	cmd.AddCommand(newGetTenantCollectionCmd().getCommand())
	cmd.AddCommand(newNewTenantCmd().getCommand())
	cmd.AddCommand(newGetTenantCmd().getCommand())
	cmd.AddCommand(newDeleteTenantCmd().getCommand())
	cmd.AddCommand(newUpdateTenantCmd().getCommand())
	cmd.AddCommand(newCurrentTenantCmd().getCommand())
	cmd.AddCommand(newEnableApplicationOnTenantCmd().getCommand())
	cmd.AddCommand(newDisableApplicationFromTenantCmd().getCommand())
	cmd.AddCommand(newGetApplicationReferenceCollectionCmd().getCommand())
	cmd.AddCommand(newGetTenantVersionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
