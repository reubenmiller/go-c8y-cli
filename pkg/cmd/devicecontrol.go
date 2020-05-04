package cmd

import (
	"github.com/spf13/cobra"
)

type deviceControlCmd struct {
	*baseCmd
}

func newDeviceControlCmd() *deviceControlCmd {
	ccmd := &deviceControlCmd{}

	cmd := &cobra.Command{
		Use:   "operation",
		Short: "Device control (operations) REST endpoint",
		Long: `
		The device control interface consists of three parts:

			* The device control API resource returns URIs and URI templates to collections of operations, so that operations can be queried by various criteria.
			* The operation collection resource retrieves operations and enables creating new operations.
			* The operation resource represents individual operations that can be queried and modified.

		In order to create/retrieve/update an operation for a device, the device must be in the “childDevices” hierarchy of an existing agent. To create an agent in the inventory, you should create a managed object with a fragment “com_cumulocity_model_Agent”.

		Note that for all PUT/POST requests accept header should be provided, otherwise an empty response body will be returned.
		`,
	}

	// Subcommands
	cmd.AddCommand(newGetOperationCmd().getCommand())
	cmd.AddCommand(newGetOperationCollectionCmd().getCommand())
	cmd.AddCommand(newNewOperationCmd().getCommand())
	cmd.AddCommand(newDeleteOperationCollectionCmd().getCommand())
	cmd.AddCommand(newUpdateOperationCmd().getCommand())

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
