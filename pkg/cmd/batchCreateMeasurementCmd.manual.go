package cmd

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/annotation"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
)

type batchCreateMeasurementCmd struct {
	*baseCmd

	source     string
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
		Annotations: map[string]string{
			annotation.FlagValueFromPipeline: "inputFile",
		},
		Example: `
$ c8y batch createMeasurements --inputList mylist.csv --template "measurement.jsonnet"
Create a measurements for a list of input devices
        `,
		PreRunE: validateBatchCreateMode,
		RunE:    ccmd.runE,
	}

	cmd.SilenceUsage = true
	// cmd.Flags().StringVar(&ccmd.source, "inputFile", "", "Input device list")
	addBatchFlags(cmd, true)
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchCreateMeasurementCmd) runE(cmd *cobra.Command, args []string) error {
	body := mapbuilder.NewMapBuilder()
	body.SetEmptyMap()
	setLazyDataTemplateFromFlags(cmd, body)
	body.Set("time", NewRelativeTimeIterator("0s"))

	sourceIter, err := NewFlagFileContents(cmd, "inputFile")
	if err != nil {
		return err
	}
	body.Set("source.id", sourceIter)
	body.TemplateIterator = iterator.NewRangeIterator(1, 1000000, 1)

	pathIter := iterator.NewRepeatIterator("measurement/measurements", 0)

	requestIter := NewBatchPathRequestIterator(
		cmd, "POST", pathIter, body)
	return runTemplateOnList(cmd, requestIter)
}
