package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/spf13/cobra"
)

type MicroservicesCmd struct {
	*subcommand.SubCommand
}

func NewMicroservicesRootCmd() *MicroservicesCmd {
	ccmd := &MicroservicesCmd{}

	cmd := &cobra.Command{
		Use:   "microservices",
		Short: "Cumulocity microservices",
		Long:  `REST endpoint to interact with Cumulocity microservices`,
	}

	// Subcommands
	cmd.AddCommand(NewGetMicroserviceCollectionCmd().GetCommand())
	cmd.AddCommand(NewGetMicroserviceCmd().GetCommand())
	cmd.AddCommand(NewDeleteMicroserviceCmd().GetCommand())
	cmd.AddCommand(NewUpdateMicroserviceCmd().GetCommand())
	cmd.AddCommand(NewNewMicroserviceBinaryCmd().GetCommand())
	cmd.AddCommand(NewGetMicroserviceBootstrapUserCmd().GetCommand())
	cmd.AddCommand(NewEnableMicroserviceCmd().GetCommand())
	cmd.AddCommand(NewDisableMicroserviceCmd().GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
