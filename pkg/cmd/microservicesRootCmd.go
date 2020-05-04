package cmd

import (
	"github.com/spf13/cobra"
)

type microservicesCmd struct {
	*baseCmd
}

func newMicroservicesRootCmd() *microservicesCmd {
	ccmd := &microservicesCmd{}

	cmd := &cobra.Command{
		Use:   "microservices",
		Short: "Cumulocity microservices",
		Long:  `REST endpoint to interact with Cumulocity microservices`,
	}

	// Subcommands
	cmd.AddCommand(newGetMicroserviceCollectionCmd().getCommand())
	cmd.AddCommand(newGetMicroserviceCmd().getCommand())
	cmd.AddCommand(newDeleteMicroserviceCmd().getCommand())
	cmd.AddCommand(newUpdateMicroserviceCmd().getCommand())
	cmd.AddCommand(newNewMicroserviceBinaryCmd().getCommand())
	cmd.AddCommand(newGetMicroserviceBootstrapUserCmd().getCommand())
	cmd.AddCommand(newEnableMicroserviceCmd().getCommand())
	cmd.AddCommand(newDisableMicroserviceCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
