package cmd

import (
	"github.com/spf13/cobra"
)

type batchCreateMeasurementCmd struct {
	*baseCmd

	count      int
	startIndex int
	workers    int
	delay      int
}

func newBatchCreateMeasurementCmd() *batchCreateMeasurementCmd {
	ccmd := &batchCreateMeasurementCmd{}

	cmd := &cobra.Command{
		Use:   "createMeasurements",
		Short: "Create a batch of measurements",
		Long:  `Create a batch of measurements`,
		Example: `
$ c8y batch createMeasurements --inputList mylist.csv --template "measurement.jsonnet"
Create a measurements for a list of input devices
        `,
		PreRunE: validateBatchCreateMode,
		RunE:    ccmd.runE,
	}

	cmd.SilenceUsage = true
	addBatchFlags(cmd, true)
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchCreateMeasurementCmd) runE(cmd *cobra.Command, args []string) error {
	return runTemplateOnList(cmd, "POST", "measurement/measurements", `{"source":{"id":"{id}"}}`)
}
