package repeatfile

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/randdata"
	"github.com/spf13/cobra"
)

type CmdRepeatFile struct {
	*subcommand.SubCommand

	infinite   bool
	format     string
	times      int64
	times_min  int64
	times_max  int64
	first      int64
	randomSkip float32
	factory    *cmdutil.Factory
}

func NewCmdRepeatFile(f *cmdutil.Factory) *CmdRepeatFile {
	ccmd := &CmdRepeatFile{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "repeatfile <FILE> [...FILE]",
		Short: "Repeat file contents",
		Long: heredoc.Doc(`
			Generic utility to repeat each line in a list of files

			The command will read the list of files and iterate through each one printing out
			each line once. The process is repeated N times (as defined by the --times flag)

			Processing order:

				=> File 1: Print lines 1 -> last
				=> File 2: Print lines 1 -> last
				=> File M: Print lines 1 -> last
				=> Repeat process X times
		`),
		Example: heredoc.Doc(`
			$ c8y util repeatfile myfile.txt
			Repeat each line of the file contents

			$ c8y util repeatfile myfile.txt --times 2
			Loop over the file twice. The file contents will be printed from first list to last line, then read again.

			$ c8y util repeatfile myfile.txt --infinite --delay 500ms
			Loop over the file contents forever and delaying 500ms after each line is printed

			$ c8y util repeatfile *.list --randomSkip 0.5
			Loop over the files matching the "*.list" in the current directory (uses shell expansion), but randomly
			skip lines at a probability of 50 percent.
		`),
		Args: cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveDefault
		},
		RunE: ccmd.newTemplate,
	}

	cmd.Flags().StringVar(&ccmd.format, "format", "%s", "format string to be applied to each input line")
	cmd.Flags().Int64Var(&ccmd.first, "first", 0, "only include first x lines. 0 = all lines")
	cmd.Flags().String("randomDelayMin", "0ms", "random minimum delay after each request, i.e. 5ms, 1.2s. It must be less than randomDelayMax. 0 = disabled")
	cmd.Flags().String("randomDelayMax", "0ms", "random maximum delay after each request, i.e. 5ms, 1.2s. It must be larger than randomDelayMin. 0 = disabled.")
	cmd.Flags().Float32Var(&ccmd.randomSkip, "randomSkip", -1, "randomly skip line based on a percentage, probability as a float: 0 to 1, 1 = always skip, 0 = never skip, -1 = disabled")
	cmd.Flags().Int64Var(&ccmd.times, "times", 1, "number of times to repeat the input")
	cmd.Flags().Int64Var(&ccmd.times_min, "min", 1, "min number of (randomized) times to repeat the input (inclusive)")
	cmd.Flags().Int64Var(&ccmd.times_max, "max", 1, "max number of (randomized) times to repeat the input (inclusive). 0 = no output")
	cmd.Flags().BoolVar(&ccmd.infinite, "infinite", false, "Repeat forever. You will need to ctrl-c it to stop it")

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

func (n *CmdRepeatFile) newTemplate(cmd *cobra.Command, args []string) error {
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

	// randomized times
	if cmd.Flags().Changed("min") || cmd.Flags().Changed("max") {
		// If only min is provided, then adjust max values to equal to min
		// This will behaviour exactly the same as using --times x.
		// However, just providing a max value will result in range from 1 to max
		if !cmd.Flags().Changed("max") {
			n.times_max = n.times_min
		}
		// Allow users to set --max 0 to disable all output
		// as it gives the user full control to also turn off the output if desired
		if n.times_max == 0 {
			times = 0
		} else {
			times = randdata.Integer(n.times_max, n.times_min)
		}
	}

	// Random delay generator
	randomDelayFunc, err := flags.GetDurationGenerator(cmd, "randomDelayMin", "randomDelayMax", true, time.Millisecond)
	if err != nil {
		return err
	}

	delayBefore := cfg.WorkerDelayBefore()
	delay := cfg.WorkerDelay()

	files := []string{}
	files = append(files, args...)
	hasInvalidPaths := false
	for _, file := range args {
		if _, err := os.Stat(file); err != nil {
			cfg.Logger.Errorf("file does not exist. path=%s. error=%s", file, err)
			hasInvalidPaths = true
		}
	}

	if hasInvalidPaths {
		return fmt.Errorf("some input files do not exist")
	}

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

	totalRows := int64(0)
	if n.first > 0 {
		totalRows = n.first
	}

	row := int64(0)
	rowCount := int64(0)
	outputCount := int64(1)
	for {
		row++

		if row > times && !n.infinite {
			break
		}

		for _, curFile := range files {
			iter, err := iterator.NewFileContentsIterator(curFile)
			if err != nil {
				return err
			}

			for {
				responseText, _, err := iter.GetNext()
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}

				if totalRows != 0 && rowCount >= totalRows {
					cfg.Logger.Debugf("Found first %d rows", rowCount)
					return nil
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

				line := ""

				if includeRowNum {
					line = fmt.Sprintf(formatString, bytes.TrimSpace(responseText), fmt.Sprintf("%d", outputCount)) + outputEnding
				} else {
					line = fmt.Sprintf(formatString, bytes.TrimSpace(responseText)) + outputEnding
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
				rowCount++
			}
		}
	}
	return nil
}
