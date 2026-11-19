// v2-based (deprecated) assign device to group: links an existing device as a
// childAsset of a group via ManagedObjects.ChildAssets.Assign. The
// newChildDevice flag drives iteration; the group is the parent. Hidden and
// deprecated in favour of `c8y devicegroups children assign --childType asset`.
package assigndevice

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

// AssignDeviceCmd command
type AssignDeviceCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignDeviceCmd creates a command to Assign device to group
func NewAssignDeviceCmd(f *cmdutil.Factory) *AssignDeviceCmd {
	ccmd := &AssignDeviceCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "assignDevice",
		Short:  "Assign device to group",
		Long:   `Assigns a device to a group. The device will be a childAsset of the group`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devicegroups assignDevice --group 12345 --newChildDevice 43234
Add a device to a group

$ c8y devicegroups assignDevice --group 12345 --newChildDevice 43234,99292,12222
Add multiple devices to a group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required)")
	cmd.Flags().StringSlice("newChildDevice", []string{""}, "New device to be added to the group as an child asset (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("newChildDevice", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("newChildDevice", "managedObject.id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Add-ChildDeviceToGroup"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children assign --childType asset' instead"),
	)

	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AssignDeviceCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("newChildDevice"); err != nil {
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
		childID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("newChildDevice")), nil)
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
