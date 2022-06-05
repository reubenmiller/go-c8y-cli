package repeat

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iterator"
	"github.com/spf13/cobra"
)

type CmdRepeat struct {
	*subcommand.SubCommand

	useTotalCount bool
	infinite      bool
	format        string
	times         int64
	skip          int64
	first         int64
	offset        int64
	randomSkip    float32
	factory       *cmdutil.Factory
}

func NewCmdRepeat(f *cmdutil.Factory) *CmdRepeat {
	ccmd := &CmdRepeat{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "repeat",
		Short: "Repeat input",
		Long:  `Generic utility to repeat input values x times`,
		Example: heredoc.Doc(`
			$ c8y util repeat 5 --input "my name"
			Repeat input value "my name" 5 times

			$ echo "my name" | c8y util repeat 2 --format "my prefix - %s"
			Repeat input value "my name" 2 times (using pipeline)
				=> my prefix - my name
				=> my prefix - my name
			
			$ echo "device" | c8y util repeat 2 --offset 100 --format "%s %05s"
			Repeat input value "device" 2 times (using pipeline)
				=> device 00101
				=> device 00102

			$ c8y util repeat --infinite | c8y api --url "/service/report-agent/health" --raw --delay 1s
			Use repeat to create an infinite loop, to check the health of a microservice waiting 1 seconds after each request

			$ echo "device" | c8y util repeat 2 | c8y util repeat 3 --format "%s_%s"
			Combine two calls to iterator over 3 devices twice. This can then be used to input into other c8y commands
				=> device_1
				=> device_2
				=> device_3
				=> device_1
				=> device_2
				=> device_3
			
			$ c8y devices get --id 1235 | c8y util repeat 5 | c8y events create --text "test event" --type "myType" --dry --delay 1000ms
			Get a device, then repeat it 5 times in order to create 5 events for it (delaying 1000 ms between each event creation)

			$ c8y devices get --id 1234 | c8y util repeat 5 --randomDelayMin 1000ms --randomDelayMax 10000ms -v | c8y events create --text "test event" --type "myType"
			Create 10 events for the same device and use a random delay between 1000ms and 10000ms between the creation of each event

			$ echo "test" | c8y util repeat 5 --randomDelayMax 10000ms -v
			Print "test" 5 times waiting between 0s and 10s after each line

			$ echo "test" | c8y util repeat 5 --randomDelayMin 5s -v
			Print "test" 5 times waiting exactly 5 seconds after each line
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: ccmd.newTemplate,
	}

	cmd.Flags().String("input", "", "input value to be repeated (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.format, "format", "%s", "format string to be applied to each input line")
	cmd.Flags().Int64Var(&ccmd.skip, "skip", 0, "skip first x input lines")
	cmd.Flags().Int64Var(&ccmd.first, "first", 0, "only include first x lines. 0 = all lines")
	cmd.Flags().Int64Var(&ccmd.offset, "offset", 0, "offset the output index counter. default = 0.")
	cmd.Flags().String("randomDelayMin", "0ms", "random minimum delay after each request, i.e. 5ms, 1.2s. It must be less than randomDelayMax. 0 = disabled")
	cmd.Flags().String("randomDelayMax", "0ms", "random maximum delay after each request, i.e. 5ms, 1.2s. It must be >= randomDelayMin. 0 = disabled.")
	cmd.Flags().Float32Var(&ccmd.randomSkip, "randomSkip", -1, "randomly skip line based on a percentage, probability as a float: 0 to 1, 1 = always skip, 0 = never skip, -1 = disabled")
	cmd.Flags().Int64Var(&ccmd.times, "times", 1, "number of times to repeat the input")
	cmd.Flags().BoolVar(&ccmd.useTotalCount, "useLineCount", false, "Use line count for the index instead of repeat counter")
	cmd.Flags().BoolVar(&ccmd.infinite, "infinite", false, "Repeat infinitely. You will need to ctrl-c it to stop it")

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("input", "input", false),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdRepeat) newTemplate(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	times := n.times
	if len(args) > 0 {
		if v, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			times = v
		}
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UTC().UnixNano())

	var iter iterator.Iterator
	_, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
		Name:        "input",
		InputFilter: func(b []byte) bool { return true },
		Disabled:    inputIterators.PipeOptions.Disabled,
		Required:    false,
	})(cmd, inputIterators)

	if err != nil {
		return cmderrors.NewUserError(err)
	}

	switch v := input.(type) {
	case iterator.Iterator:
		iter = v
	default:
		// use a single input iterator
		iter = iterator.NewRepeatIterator("", 1)
	}

	// Random delay generator
	randomDelayFunc, err := flags.GetDurationGenerator(cmd, "randomDelayMin", "randomDelayMax", true, time.Millisecond)
	if err != nil {
		return err
	}

	delayBefore := cfg.WorkerDelayBefore()
	delay := cfg.WorkerDelay()

	formatString := n.format

	refs := strings.Count(formatString, "%")
	if refs == 0 {
		formatString += "%s"
	}
	includeRowNum := false

	if refs > 1 {
		includeRowNum = true
	}

	cfg.Logger.Infof("repeat format string: %s", formatString)

	firstRow := int64(0)
	if n.skip > 0 {
		firstRow = n.skip
	}

	totalRows := int64(0)
	if n.first > 0 {
		totalRows = n.first
	}

	bounded := iter.IsBound()
	row := int64(0)
	rowCount := int64(0)
	outputCount := int64(1)
	for {
		row++
		responseText, _, err := iter.GetNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if firstRow != 0 && row <= firstRow {
			cfg.Logger.Debugf("Skipping row: %d", row)
			continue
		}

		if totalRows != 0 && rowCount >= totalRows {
			cfg.Logger.Debugf("Found first %d rows", rowCount)
			break
		}

		if n.randomSkip >= -1 {
			// randomly skip a row. 1 = always skip, 0 = never skip
			randValue := rand.Float32()
			if randValue <= n.randomSkip {
				cfg.Logger.Debugf("Skipping random row: %d. value=%f, limit=%f", row, randValue, n.randomSkip)
				continue
			}
		}

		outputEnding := "\n"

		// Using space if no value is provided so the piped data does not get ignored
		// by downstream commands
		if len(responseText) == 0 {
			outputEnding = " " + outputEnding
		}

		for i := int64(0); i < times || n.infinite; i++ {
			line := ""

			if includeRowNum {
				index := 1 + (outputCount-1)%times
				if n.useTotalCount {
					index = outputCount
				}
				line = fmt.Sprintf(formatString, responseText, fmt.Sprintf("%d", index+n.offset)) + outputEnding
			} else {
				line = fmt.Sprintf(formatString, responseText) + outputEnding
			}

			if delayBefore > 0 {
				time.Sleep(delayBefore)
			}

			fmt.Print(line)

			currentDelay := randomDelayFunc(delay)
			if currentDelay > 0 {
				cfg.Logger.Infof("Waiting %v before printing next value", currentDelay)
				time.Sleep(currentDelay)
			}
			outputCount++
		}

		if !bounded {
			break
		}

		rowCount++
	}

	return nil
}
