// v2-based (deprecated) unassign device from group: removes a device's
// childAsset reference from a group via ManagedObjects.ChildAssets.Unassign. The
// childDevice flag drives iteration; the group is the parent. Hidden and
// deprecated in favour of `c8y devicegroups children unassign --childType asset`.
package unassigndevice

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

// UnassignDeviceCmd command
type UnassignDeviceCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUnassignDeviceCmd creates a command to Unassign device from group
func NewUnassignDeviceCmd(f *cmdutil.Factory) *UnassignDeviceCmd {
	ccmd := &UnassignDeviceCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "unassignDevice",
		Short:  "Unassign device from group",
		Long:   `Unassign/delete a device from a group`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devicegroups unassignDevice --group 12345 --childDevice 22553
Unassign a child device from its parent group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Asset id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("childDevice", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("childDevice", "reference", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-DeviceFromGroup"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children unassign --childType asset' instead"),
	)

	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnassignDeviceCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("childDevice"); err != nil {
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
		childID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("childDevice")), nil)
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
