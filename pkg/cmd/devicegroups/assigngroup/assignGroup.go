// v2-based (deprecated) assign child group: links an existing group as a
// childAsset of a group via ManagedObjects.ChildAssets.Assign. The
// newChildGroup flag drives iteration; the group is the parent. Hidden and
// deprecated in favour of `c8y devicegroups children assign --childType asset`.
package assigngroup

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

// AssignGroupCmd command
type AssignGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignGroupCmd creates a command to Assign child group
func NewAssignGroupCmd(f *cmdutil.Factory) *AssignGroupCmd {
	ccmd := &AssignGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "assignGroup",
		Short:  "Assign child group",
		Long:   `Assigns a group to a group. The group will be a childAsset of the group`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devicegroups assignGroup --group 12345 --newChildGroup 43234
Add a group to a group

$ c8y devicegroups assignGroup --group 12345 --newChildGroup 43234,99292,12222
Add multiple groups to a group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required)")
	cmd.Flags().StringSlice("newChildGroup", []string{""}, "New child group to be added to the group as an child asset (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDeviceGroup("newChildGroup", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("newChildGroup", "managedObject.id", true, "id"),
		flags.WithPowershellName("Add-ChildGroupToGroup"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children unassign --childType asset' instead"),
	)

	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AssignGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("newChildGroup"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		groupID, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(firstValue(in.StringSlice("group"))), nil)
		if err != nil {
			return nil, err
		}
		childID, err := client.DeviceGroups.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("newChildGroup")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.ChildAssets.Assign(ctx, groupID, childID)
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
