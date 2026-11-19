// v2-based (deprecated) assign child device: links an existing device as a
// childDevice of a parent device via ManagedObjects.ChildDevices.Assign. The
// newChild flag drives iteration; the device is the parent. Hidden and
// deprecated in favour of `c8y devices children assign --childType device`.
package assignchild

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

// AssignChildCmd command
type AssignChildCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignChildCmd creates a command to Assign child device
func NewAssignChildCmd(f *cmdutil.Factory) *AssignChildCmd {
	ccmd := &AssignChildCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "assignChild",
		Short:  "Assign child device",
		Long:   `Create a child device reference`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devices assignChild --device 12345 --newChild 44235
Assign a device as a child device to an existing device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device. (required)")
	cmd.Flags().StringSlice("newChild", []string{""}, "New child device (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("newChild", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("newChild", "managedObject.id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("newChild", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Add-ChildDeviceToDevice"),
		flags.WithDeprecationNotice("please use 'c8y devices children assign --childType device' instead"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("device")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AssignChildCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("newChild"); err != nil {
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
		childID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("newChild")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.ChildDevices.Assign(ctx, deviceID, childID)
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
