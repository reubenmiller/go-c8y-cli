// v2-based notification2 delete-by-source: the device flag drives iteration and
// is resolved (name -> id) by go-c8y into the source query parameter before the
// collection delete.
package deletebysource

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/notification2"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteBySourceCmd command
type DeleteBySourceCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteBySourceCmd creates a command to Delete subscription by source
func NewDeleteBySourceCmd(f *cmdutil.Factory) *DeleteBySourceCmd {
	ccmd := &DeleteBySourceCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "deleteBySource",
		Short: "Delete subscription by source",
		Long:  `Delete an existing subscription associated to a managed object`,
		Example: heredoc.Doc(`
$ c8y notification2 subscriptions deleteBySource --device 12345
Delete a subscription associated with a device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The managed object to which the subscription is associated. (accepts pipeline)")
	cmd.Flags().String("context", "", "The context to which the subscription is associated.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("context", "mo", "tenant"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-Notification2SubscriptionBySource"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteBySourceCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := notification2.DeleteBySourceOptions{
			Context: in.String("context"),
		}
		if device := in.String("device"); device != "" {
			opt.Source = c8ystream.NameOrID(device)
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Notification2.DeleteBySource(ctx, opt)
			})
		}, nil
	})
}
