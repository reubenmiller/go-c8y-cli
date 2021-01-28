package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type valueGenerator struct {
	start   int64
	end     int64
	step    int64
	current int64
}

func newValueGenerator(start, end, step int64) *valueGenerator {
	return &valueGenerator{
		start:   start,
		end:     end,
		step:    step,
		current: start - 1,
	}
}

func (p *valueGenerator) ReadBytes(delim byte) (line []byte, err error) {
	nextValue := atomic.AddInt64(&p.current, p.step)
	if nextValue > p.end {
		err = io.EOF
	} else {
		line = []byte(fmt.Sprintf("%d%c", nextValue, delim))
	}
	return line, err
}

func newPipeReader() (InputIterator, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	// if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
	if info.Mode()&os.ModeCharDevice != 0 {
		return nil, errors.New("No pipe line input detected")
	}

	reader := bufio.NewReader(os.Stdin)

	return reader, nil
}

func newFileReader(path string) (*bufio.Reader, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(fp), nil
}

type batchDummyPipeCmd struct {
	*baseCmd

	inputFile string
}

type InputIterator interface {
	ReadBytes(delim byte) (line []byte, err error)
	//ReadString(delim byte) (line string, err error)
}

func newBatchDummyPipeCmd() *batchDummyPipeCmd {
	ccmd := &batchDummyPipeCmd{}

	cmd := &cobra.Command{
		Use:   "dummy",
		Short: "Dummy command to test piped data",
		Long:  `Dummy command to test piped data`,
		Annotations: map[string]string{
			"pipedArg": "inputFile",
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

	var buf InputIterator
	var err error

	buf, err = newPipeReader()
	if err != nil {
		Logger.Debug("No pipeline input detected")
		if cmd.Flags().Changed("inputFile") {
			fmt.Println("Reading from file")
			// fill data from file
			pathValue, err := cmd.Flags().GetString("inputFile")
			if err != nil {
				return err
			}
			buf, err = newFileReader(pathValue)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Printf("PIPED Input:\n")
	}

	// buf = newValueGenerator(1, 10, 1)

	// if cmd.Flags().Changed("inputFile") {
	// 	buf, err = newPipeReader()
	// 	if err != nil {
	// 		Logger.Debug("No pipeline input detected")
	// 	} else {
	// 		fmt.Printf("PIPED Input:\n")
	// 	}
	// }

	// path := fmt.Sprintf("inventory/managedObjects/%s/childAssets", n.group)
	// return runTemplateOnList(cmd, "POST", path, `{"managedObject":{"id":"{id}"}}`)
	return n.processE(cmd, buf)
}

func (n *batchDummyPipeCmd) processE(cmd *cobra.Command, input InputIterator) error {
	for {
		line, err := input.ReadBytes('\n')
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
