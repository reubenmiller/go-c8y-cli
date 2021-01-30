package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/reubenmiller/go-c8y-cli/pkg/annotation"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/spf13/cobra"
)

type batchDummyPipeCmd struct {
	*baseCmd

	inputFile string
}

func newBatchDummyPipeCmd() *batchDummyPipeCmd {
	ccmd := &batchDummyPipeCmd{}

	cmd := &cobra.Command{
		Use:   "dummy",
		Short: "Dummy command to test piped data",
		Long:  `Dummy command to test piped data`,
		Annotations: map[string]string{
			annotation.FlagValueFromPipeline: "inputFile",
		},
		Example: `
$ ls -l | c8y batch dummy
Pipe a list of files to c8y
        `,
		// PreRunE: validateBatchCreateMode,
		RunE: ccmd.runE,
	}

	cmd.SilenceUsage = true
	cmd.Flags().StringVar(&ccmd.inputFile, "inputFile", "", "Group")
	//addBatchFlags(cmd, true)
	//addDataFlag(cmd)
	//addProcessingModeFlag(cmd)

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *batchDummyPipeCmd) runE(cmd *cobra.Command, args []string) error {

	iter, err := FlagGetFileIterator(cmd, "inputFile")
	if err != nil {
		return err
	}

	// iter = iterator.NewRangeIterator(1, 100, 1)

	// path := fmt.Sprintf("inventory/managedObjects/%s/childAssets", n.group)
	// return runTemplateOnList(cmd, "POST", path, `{"managedObject":{"id":"{id}"}}`)
	return n.processE(cmd, iter)
}

func (n *batchDummyPipeCmd) processE(cmd *cobra.Command, iter iterator.Iterator) error {
	data := make(map[string]interface{}, 0)
	data["test"] = iter
	if out, err := json.Marshal(data); err == nil {
		fmt.Printf("Iter to json: %s\n", out)
	} else {
		fmt.Printf("json error: %s\n", err)
	}
	for {
		line, err := iter.GetNext()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		fmt.Printf("line: %s", line)
	}
	fmt.Println("input finished")
	return nil
}

// FlagGetStringIterator returns an iterator from a command flag
func FlagGetStringIterator(cmd *cobra.Command, name string) (iterator.Iterator, error) {
	var iter iterator.Iterator
	var err error

	iter = &iterator.EmptyIterator{}

	if cmd.Flags().Changed(name) {
		// user is using flags (ignore piped input)
		Logger.Info("Reading from arguments")

		values, err := cmd.Flags().GetStringSlice(name)
		if err != nil {
			return nil, err
		}
		iter = iterator.NewSliceIterator(values)
	} else {
		// check stdin for data
		iter, err = iterator.NewPipeIterator()
		if err != nil && err != iterator.ErrNoPipeInput {
			return nil, err
		}
	}
	return iter, nil
}

// FlagGetFileIterator returns an iterator from a command flag
func FlagGetFileIterator(cmd *cobra.Command, name string) (iterator.Iterator, error) {
	var iter iterator.Iterator
	var err error

	iter = &iterator.EmptyIterator{}

	if cmd.Flags().Changed(name) {
		// user is using flags (ignore piped input)
		Logger.Info("Reading from file")

		pathValue, err := cmd.Flags().GetString(name)
		if err != nil {
			return nil, err
		}
		iter, err = iterator.NewFileContentsIterator(pathValue)
		if err != nil {
			return nil, err
		}
	} else {
		// check stdin for data
		iter, err = iterator.NewPipeIterator()
		if err != nil && err != iterator.ErrNoPipeInput {
			return nil, err
		}
	}
	return iter, nil
}
