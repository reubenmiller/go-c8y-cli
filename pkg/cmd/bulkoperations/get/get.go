// v2-based bulk operations get: the id flag drives iteration (pipe or --id
// slice); each bulk operation is fetched via BulkOperations.Get by id.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get bulk operation
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get bulk operation",
		Long:  `Get an existing bulk operation`,
		Example: heredoc.Doc(`
$ c8y bulkoperations get --id 12345
Get bulk operation by id
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Bulk Operation id (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Get-BulkOperation"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.bulkoperation+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id := in.String("id")
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.BulkOperations.Get(ctx, id))
		}, nil
	})
}
