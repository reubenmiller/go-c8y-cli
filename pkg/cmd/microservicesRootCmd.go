package cmd

import (
	"github.com/spf13/cobra"
)

type MicroservicesCmd struct {
	*baseCmd
}

func NewMicroservicesRootCmd() *MicroservicesCmd {
	ccmd := &MicroservicesCmd{}

	cmd := &cobra.Command{
		Use:   "microservices",
		Short: "Cumulocity microservices",
		Long:  `REST endpoint to interact with Cumulocity microservices`,
	}

	// Subcommands
	cmd.AddCommand(NewGetMicroserviceCollectionCmd().getCommand())
	cmd.AddCommand(NewGetMicroserviceCmd().getCommand())
	cmd.AddCommand(NewDeleteMicroserviceCmd().getCommand())
	cmd.AddCommand(NewUpdateMicroserviceCmd().getCommand())
	cmd.AddCommand(NewNewMicroserviceBinaryCmd().getCommand())
	cmd.AddCommand(NewGetMicroserviceBootstrapUserCmd().getCommand())
	cmd.AddCommand(NewEnableMicroserviceCmd().getCommand())
	cmd.AddCommand(NewDisableMicroserviceCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
