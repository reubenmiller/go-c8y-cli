package cmd

import (
	"github.com/spf13/cobra"
)

type batchUpdateManagedObjectCmd struct {
	*baseCmd

	group string
}

func newBatchUpdateManagedObjectCmd() *batchUpdateManagedObjectCmd {
	ccmd := &batchUpdateManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "updateManagedObjects",
		Short: "Update a list of managed objects",
		Long:  `Update a list of managed objects by accepting a list of managed object ids and applying the given template to them`,
		Example: `
$ c8y batch updateManagedObjects --inputList mylist.csv --template "update.template.jsonnet"
Update a list of managed objects

$ c8y batch updateManagedObjects --inputList mylist.csv --template "update.template.jsonnet" --dry
Do a dry-run by only showing the requests on console to check that the commands are correct
        `,
		PreRunE: validateBatchDeleteMode,
		RunE:    ccmd.runE,
	}

	cmd.SilenceUsage = true

	addBatchFlags(cmd, true)
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchUpdateManagedObjectCmd) runE(cmd *cobra.Command, args []string) error {
	return runTemplateOnList(cmd, "PUT", "inventory/managedObjects/{id}", "")
}
