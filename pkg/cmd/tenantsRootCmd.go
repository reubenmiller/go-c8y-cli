package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type TenantsCmd struct {
	*subcommand.SubCommand
}

func NewTenantsRootCmd() *TenantsCmd {
	ccmd := &TenantsCmd{}

	cmd := &cobra.Command{
		Use:   "tenants",
		Short: "Cumulocity tenant",
		Long:  `REST endpoint to interact with Cumulocity tenants`,
	}

	// Subcommands
	cmd.AddCommand(NewGetTenantCollectionCmd().GetCommand())
	cmd.AddCommand(NewNewTenantCmd().GetCommand())
	cmd.AddCommand(NewGetTenantCmd().GetCommand())
	cmd.AddCommand(NewDeleteTenantCmd().GetCommand())
	cmd.AddCommand(NewUpdateTenantCmd().GetCommand())
	cmd.AddCommand(NewCurrentTenantCmd().GetCommand())
	cmd.AddCommand(NewEnableApplicationOnTenantCmd().GetCommand())
	cmd.AddCommand(NewDisableApplicationFromTenantCmd().GetCommand())
	cmd.AddCommand(NewGetApplicationReferenceCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetTenantVersionCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
