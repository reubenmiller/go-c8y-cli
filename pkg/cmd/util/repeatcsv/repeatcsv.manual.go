package fromcsv

import (
	"fmt"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/spf13/cobra"
)

type CmdRepeatCsvFile struct {
	*subcommand.SubCommand

	noHeader   bool
	delimiter  string
	columns    []string
	infinite   bool
	times      int64
	first      int64
	randomSkip float32
	factory    *cmdutil.Factory
}

func (c *CmdRepeatCsvFile) HasHeader() bool {
	return !c.noHeader
}

func NewCmdFromCsv(f *cmdutil.Factory) *CmdRepeatCsvFile {
	ccmd := &CmdRepeatCsvFile{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "repeatcsv <FILE> [...FILE]",
		Short: "Iterate over csv files and convert them to json lines",
		Long: heredoc.Doc(`
			Generic utility to iterate over csv files and transform them into newline delimited json objects
			which can be piped to other commands.
		`),
		Example: heredoc.Doc(`
			$ c8y util repeatcsv input.csv
			Convert the input.csv to json

			$ c8y util repeatcsv *.csv
			Convert multiple csv files

			$ c8y util repeatcsv input.csv --columns custom1,custom2,custom3
			Convert csv input using custom header
		`),
		Args: cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveDefault
		},
		RunE: ccmd.newTemplate,
	}

	cmd.Flags().StringVar(&ccmd.delimiter, "delimiter", "", "Field delimiter. It will be auto detected by default.")
	cmd.Flags().BoolVar(&ccmd.noHeader, "noHeader", false, "Input data does not have a header row")
	cmd.Flags().StringSliceVar(&ccmd.columns, "columns", []string{}, "Columns")
	cmd.Flags().Int64Var(&ccmd.first, "first", 0, "only include first x lines. 0 = all lines")
	cmd.Flags().String("randomDelayMin", "0ms", "random minimum delay after each request, i.e. 5ms, 1.2s. It must be less than randomDelayMax. 0 = disabled")
	cmd.Flags().String("randomDelayMax", "0ms", "random maximum delay after each request, i.e. 5ms, 1.2s. It must be larger than randomDelayMin. 0 = disabled.")
	cmd.Flags().Float32Var(&ccmd.randomSkip, "randomSkip", -1, "randomly skip line based on a percentage, probability as a float: 0 to 1, 1 = always skip, 0 = never skip, -1 = disabled")
	cmd.Flags().Int64Var(&ccmd.times, "times", 1, "number of times to repeat the input")
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

func (n *CmdRepeatCsvFile) newTemplate(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

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

	iterFactory := func(path string) (iterator.Iterator, error) {
		return iterator.NewCSVFileContentsIterator(path, n.delimiter, n.HasHeader(), n.columns)
	}

	return cmdutil.ExecuteFileIterator(n.GetCommand().OutOrStdout(), cfg.Logger, files, iterFactory, cmdutil.FileIteratorOptions{
		Infinite:        n.infinite,
		Times:           n.times,
		FirstNRows:      n.first,
		RandomDelayFunc: randomDelayFunc,
		Delay:           delay,
		DelayBefore:     delayBefore,
	})
}
