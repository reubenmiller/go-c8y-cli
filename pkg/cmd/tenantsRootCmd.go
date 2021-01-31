package cmd

import (
	"github.com/spf13/cobra"
)

type TenantsCmd struct {
	*baseCmd
}

func NewTenantsRootCmd() *TenantsCmd {
	ccmd := &TenantsCmd{}

	cmd := &cobra.Command{
		Use:   "tenants",
		Short: "Cumulocity tenant",
		Long:  `REST endpoint to interact with Cumulocity tenants`,
	}

	// Subcommands
	cmd.AddCommand(NewGetTenantCollectionCmd().getCommand())
	cmd.AddCommand(NewNewTenantCmd().getCommand())
	cmd.AddCommand(NewGetTenantCmd().getCommand())
	cmd.AddCommand(NewDeleteTenantCmd().getCommand())
	cmd.AddCommand(NewUpdateTenantCmd().getCommand())
	cmd.AddCommand(NewCurrentTenantCmd().getCommand())
	cmd.AddCommand(NewEnableApplicationOnTenantCmd().getCommand())
	cmd.AddCommand(NewDisableApplicationFromTenantCmd().getCommand())
	cmd.AddCommand(NewGetApplicationReferenceCollectionCmd().getCommand())
	cmd.AddCommand(NewGetTenantVersionCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
