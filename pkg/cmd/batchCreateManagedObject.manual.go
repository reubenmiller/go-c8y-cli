package cmd

import (
	"github.com/spf13/cobra"
)

type batchCreateManagedObjectCmd struct {
	*baseCmd

	count      int
	startIndex int
	workers    int
	delay      int
}

func newBatchCreateManagedObjectCmd() *batchCreateManagedObjectCmd {
	ccmd := &batchCreateManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "createManagedObjects",
		Short: "Create a batch of managed objects",
		Long:  `Create a batch of managed objects`,
		Example: `
$ c8y batch createManagedObjects --name "testMO" --type "custom_type"
Create a managed object
        `,
		PreRunE: validateBatchCreateMode,
		RunE:    ccmd.runE,
	}

	cmd.SilenceUsage = true
	addBatchFlags(cmd, false)
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchCreateManagedObjectCmd) runE(cmd *cobra.Command, args []string) error {
	return runTemplateOnList(cmd, "POST", "inventory/managedObjects", "")
}
