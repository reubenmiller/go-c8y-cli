package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type batchAddDeviceToGroupCmd struct {
	*baseCmd

	group string
}

func newBatchAddDeviceToGroupCmd() *batchAddDeviceToGroupCmd {
	ccmd := &batchAddDeviceToGroupCmd{}

	cmd := &cobra.Command{
		Use:   "addChildDevices",
		Short: "Create child devices to group",
		Long:  `Create child devices to group`,
		Example: `
$ c8y batch addChildDevices --group 1234 --childList mylist.csv
Add list of children to a group
        `,
		PreRunE: validateBatchCreateMode,
		RunE:    ccmd.runE,
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringVar(&ccmd.group, "group", "", "Group (required)")
	addBatchFlags(cmd, true)
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("group")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchAddDeviceToGroupCmd) runE(cmd *cobra.Command, args []string) error {
	path := fmt.Sprintf("inventory/managedObjects/%s/childAssets", n.group)
	return runTemplateOnList(cmd, "POST", path, `{"managedObject":{"id":"{id}"}}`)
}
