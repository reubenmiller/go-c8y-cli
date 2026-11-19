// v2-based operations cancel: the id flag drives iteration (pipe or --id slice);
// each operation is cancelled by writing a new status (with a failure reason)
// via Operations.Update. Cancelling is just a status update, so the id is passed
// through unresolved.
package cancel

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CancelCmd command
type CancelCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCancelCmd creates a command to Cancel operation
func NewCancelCmd(f *cmdutil.Factory) *CancelCmd {
	ccmd := &CancelCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel operation",
		Long:  `Cancel an existing operation by setting its status (and a failure reason)`,
		Example: heredoc.Doc(`
$ c8y operations cancel --id 12345
Cancel an operation
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Operation id (required) (accepts pipeline)")
	cmd.Flags().String("status", "FAILED", "Operation status")
	cmd.Flags().String("failureReason", "User cancelled operation", "Reason for the failure")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Invoke-CancelOperation"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operation+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CancelCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("failureReason", "failureReason"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("status"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id := in.String("id")
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Operation] {
				return client.Operations.Update(ctx, id, body)
			})
		}, nil
	})
}
