package cmd

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/annotation"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
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
		Annotations: map[string]string{
			annotation.FlagValueFromPipeline: "inputFile",
		},
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
	body := mapbuilder.NewMapBuilder()

	// idIter, err := FlagGetStringIterator(cmd, "inputFile")
	idIter, err := NewFlagFileContents(cmd, "inputFile")
	if err != nil {
		return err
	}
	if err := body.Set("managedObject.id", idIter); err != nil {
		return err
	}

	// NewFlagPipeEnabledStringSlice(cmd, "inputFile")
	requestIter := NewBatchFixedPathRequestIterator(cmd, "POST", path, body)
	return runTemplateOnList(cmd, requestIter)
}
