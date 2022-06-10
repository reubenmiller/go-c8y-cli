package repeat

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/spf13/cobra"
)

type CmdShow struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdShow(f *cmdutil.Factory) *CmdShow {
	ccmd := &CmdShow{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show json input",
		Long:  `Generic utility to process json lines (one json object per line) using the same logic used in other c8y commands`,
		Example: heredoc.Doc(`
			$ echo myfile.json | c8y util show --select id,name
			Process input json lines files and select id and name fields

			$ c8y devices list > devices.json
			$ c8y util show --input devices.json --select id,name --output csv
			Save a devices list to file, then process the file in a second step and convert it to csv only keeping id and name columns (with no headers)  
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().String("input", "", "input value to be repeated (required) (accepts pipeline)")

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("input", "input", true),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdShow) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UTC().UnixNano())

	var iter iterator.Iterator
	_, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
		Name:        "input",
		InputFilter: flags.FilterJsonLines,
		Disabled:    inputIterators.PipeOptions.Disabled,
		Required:    true,
	})(cmd, inputIterators)

	if err != nil {
		return &flags.ParameterError{
			Name: "input",
			Err:  fmt.Errorf("Missing required parameter or pipeline input. %w", flags.ErrParameterMissing),
		}
	}

	switch v := input.(type) {
	case iterator.Iterator:
		iter = v
	default:
		// use a single input iterator
		iter = iterator.NewRepeatIterator("", 1)
	}

	bounded := iter.IsBound()
	for {
		responseText, _, err := iter.GetNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := n.factory.WriteJSONToConsole(cfg, cmd, "", responseText); err != nil {
			cfg.Logger.Warnf("Could not process line. only json lines are accepted. %s", err)
		}

		if !bounded {
			break
		}
	}

	return nil
}
