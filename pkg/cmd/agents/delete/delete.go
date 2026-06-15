// v2-based agent delete: the id flag drives iteration (pipe or --id); each agent
// reference (id or name, resolved scoped to agents) is deleted via Agents.Delete
// along with its data. Success yields no output.
package delete

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/agents"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCmd creates a command to Delete agent
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete agent",
		Long: heredoc.Doc(`Delete an agent from the platform. This will delete all data associated to the agent
(i.e. alarms, events, operations and measurements)
`),
		Example: heredoc.Doc(`
$ c8y agents delete --id 12345
Delete agent by id

$ c8y agents delete --id 12345 --withDeviceUser
Delete agent and related device user/credentials
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Agent ID (required) (accepts pipeline)")
	cmd.Flags().Bool("withDeviceUser", false, "Delete associated device owner")
	cmd.Flags().Bool("cascade", false, "Remove all child devices and child assets will be deleted recursively. By default, the delete operation is propagated to the subgroups only if the deleted object is a group")

	completion.WithOptions(
		cmd,
		completion.WithAgent("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Remove-Agent"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCmd) RunE(cmd *cobra.Command, args []string) error {
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
		ref := agents.AgentRef(c8ystream.NameOrID(in.String("id")))
		opt := agents.DeleteOptions{
			Cascade:        in.Bool("cascade"),
			WithDeviceUser: in.Bool("withDeviceUser"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Agents.Delete(ctx, ref, opt)
			})
		}, nil
	})
}
