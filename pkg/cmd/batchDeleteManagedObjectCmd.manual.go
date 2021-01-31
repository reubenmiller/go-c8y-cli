package cmd

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/spf13/cobra"
)

type batchDeleteManagedObjectCmd struct {
	*baseCmd

	group string
}

func newBatchDeleteManagedObjectCmd() *batchDeleteManagedObjectCmd {
	ccmd := &batchDeleteManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "deleteManagedObjects",
		Short: "Delete a list of managed objects",
		Long:  `Delete a list of managed objects by accepting a list of managed object ids`,
		Example: `
$ c8y batch deleteManagedObjects --inputList mylist.csv --workers 5
Delete a list of managed objects using 5 workers
		`,
		PreRunE: validateBatchDeleteMode,
		RunE:    ccmd.runE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithBatchOptions(true),
		flags.WithProcessingMode(),
		flags.WithPipelineSupport("inputFile"),
	)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchDeleteManagedObjectCmd) runE(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not implemented")
	// return runTemplateOnList(cmd, "DELETE", "inventory/managedObjects/{id}", "")
}
