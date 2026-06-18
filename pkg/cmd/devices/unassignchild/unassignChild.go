// v2-based (deprecated) unassign child device: removes a child device reference
// from a parent device via ManagedObjects.ChildDevices.Unassign. The childDevice
// flag drives iteration; the device is the parent. Hidden and deprecated in
// favour of `c8y devices children unassign --childType device`.
package unassignchild

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

// UnassignChildCmd command
type UnassignChildCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUnassignChildCmd creates a command to Delete child device reference
func NewUnassignChildCmd(f *cmdutil.Factory) *UnassignChildCmd {
	ccmd := &UnassignChildCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "unassignChild",
		Short:  "Delete child device reference",
		Long:   `Delete child device reference`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devices unassignChild --device 12345 --childDevice 22553
Unassign a child device from its parent device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "ManagedObject id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device reference (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("childDevice", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("childDevice", "childDevice", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("childDevice", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-ChildDeviceFromDevice"),
		flags.WithDeprecationNotice("please use 'c8y devices children unassign --childType device' instead"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("device")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnassignChildCmd) RunE(cmd *cobra.Command, args []string) error {
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
		deviceID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(firstValue(in.StringSlice("device"))), nil)
		if err != nil {
			return nil, err
		}
		childID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("childDevice")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.ChildDevices.Unassign(ctx, deviceID, childID)
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
