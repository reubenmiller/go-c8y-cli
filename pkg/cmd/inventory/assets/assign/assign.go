// v2-based inventory assets assign (deprecated; use `inventory children assign
// --childType asset`): assigns a device (--childDevice) or group (--childGroup) to
// a parent group as a child asset. The childDevice flag drives iteration; the
// child reference (device or group, resolved by name or id) is linked via the
// ChildAssets service. Success yields no output.
package assign

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/children/childref"
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

// AssignCmd command
type AssignCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignCmd creates a command to Assign child asset
func NewAssignCmd(f *cmdutil.Factory) *AssignCmd {
	ccmd := &AssignCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign child asset",
		Long:  `Assigns a group or device to an existing group and marks them as assets`,
		Example: heredoc.Doc(`
$ c8y inventory assets assign --id 12345 --childGroup 43234
Create group hierarchy (parent group -> child group)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Managed object id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "New child device to be added to the group as an asset (accepts pipeline)")
	cmd.Flags().StringSlice("childGroup", []string{""}, "New child device group to be added to the group as an asset")

	completion.WithOptions(
		cmd,
		completion.WithDevice("childDevice", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDeviceGroup("childGroup", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("childDevice", "managedObject.id", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("childDevice", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("childGroup", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("managedObject"),
		flags.WithPowershellName("Add-ChildAssetToManagedObject"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReference+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
	)

	_ = cmd.MarkFlagRequired("id")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AssignCmd) RunE(cmd *cobra.Command, args []string) error {
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
		childRef := in.String("childDevice")
		if childRef == "" {
			childRef = childref.First(in.StringSlice("childGroup"))
		}
		if childRef == "" {
			return nil, fmt.Errorf("either childDevice or childGroup is required")
		}
		childID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(childRef), nil)
		if err != nil {
			return nil, err
		}
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(childref.First(in.StringSlice("id"))), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.ChildAssets.Assign(ctx, parentID, childID)
			})
		}, nil
	})
}
