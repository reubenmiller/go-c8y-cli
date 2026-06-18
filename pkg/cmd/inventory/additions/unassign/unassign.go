// v2-based inventory additions unassign (deprecated; use `inventory children
// unassign --childType addition`): removes a child-addition reference from a
// parent (resolved as a managed object) via DELETE …/childAdditions/{child}. The
// child flag drives iteration; success yields no output.
package unassign

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/children/childref"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UnassignCmd command
type UnassignCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUnassignCmd creates a command to Unassign child addition
func NewUnassignCmd(f *cmdutil.Factory) *UnassignCmd {
	ccmd := &UnassignCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "unassign",
		Short: "Unassign child addition",
		Long:  `Unassign/delete a child addition from an existing managed object`,
		Example: heredoc.Doc(`
$ c8y inventory additions unassign --id 12345 --child 22553
Unassign a child addition from its parent managed object
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Managed object id (required)")
	cmd.Flags().String("child", "", "Child managed object id (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("child", "child", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Remove-ChildAddition"),
	)

	_ = cmd.MarkFlagRequired("id")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnassignCmd) RunE(cmd *cobra.Command, args []string) error {
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
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(childref.First(in.StringSlice("id"))), nil)
		if err != nil {
			return nil, err
		}
		childID := in.String("child")
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.ManagedObjects.ChildAdditions.Unassign(ctx, parentID, childID)
			})
		}, nil
	})
}
