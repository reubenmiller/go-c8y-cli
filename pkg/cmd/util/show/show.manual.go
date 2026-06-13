// v2-based util show: a passthrough that runs each JSON-line input item
// through the shared output pipeline (--select / --filter / --output /
// --outputFile), without any API call. Used heavily in tests to inspect and
// validate input processing and the dry-run request output.
package repeat

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
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
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputRaw("input"); err != nil {
		return &flags.ParameterError{
			Name: "input",
			Err:  flags.ErrParameterMissing,
		}
	}

	// No client: each JSON-object line is yielded as-is into the pipeline.
	// Non-object lines are skipped (matching the v1 behaviour).
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		line := in.Input()
		return func(context.Context) output.Seq {
			if !jsonUtilities.IsJSONObject(line) {
				return c8ystream.FromDocs()
			}
			return c8ystream.FromDocs(jsondoc.New(line))
		}, nil
	})
}
