// v2-based (deprecated) delete child group reference: removes a group's
// childAsset reference from a parent group via ManagedObjects.ChildAssets.Unassign.
// The child flag drives iteration; the id is the parent. Hidden and deprecated
// in favour of `c8y devicegroups children unassign --childType asset`.
package unassigngroup

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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UnassignGroupCmd command
type UnassignGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUnassignGroupCmd creates a command to Delete child group reference
func NewUnassignGroupCmd(f *cmdutil.Factory) *UnassignGroupCmd {
	ccmd := &UnassignGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "unassignGroup",
		Short:  "Delete child group reference",
		Long:   `Delete child group reference`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devicegroups unassignGroup --id 12345 --child 22553
Unassign a child group from its parent group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device group (required)")
	cmd.Flags().StringSlice("child", []string{""}, "Child device group (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDeviceGroup("child", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("child", "child", true, "id"),
		flags.WithPowershellName("Remove-GroupFromGroup"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children unassign --childType asset' instead"),
	)

	_ = cmd.MarkFlagRequired("id")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnassignGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("child"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		groupID, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(firstValue(in.StringSlice("id"))), nil)
		if err != nil {
			return nil, err
		}
		childID, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("child")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.ChildAssets.Unassign(ctx, groupID, childID)
			})
		}, nil
	})
}

// firstValue returns the first non-empty entry of a slice flag (the path
// parameter takes a single value even though the flag is a slice).
func firstValue(values []string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
