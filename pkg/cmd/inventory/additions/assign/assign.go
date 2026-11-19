// v2-based inventory additions assign (deprecated; use `inventory children assign
// --childType addition`): links an existing managed object as a child addition to
// a parent (resolved as a managed object). The child flag drives iteration;
// success yields no output.
package assign

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

// AssignCmd command
type AssignCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignCmd creates a command to Assign child addition
func NewAssignCmd(f *cmdutil.Factory) *AssignCmd {
	ccmd := &AssignCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign child addition",
		Long:  `Add an existing managed object as a child addition to another existing managed object`,
		Example: heredoc.Doc(`
$ c8y inventory additions assign --id 12345 --child 6789
Add a related managed object as a child to an existing managed object
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Managed object id where the child addition will be added to (required)")
	cmd.Flags().String("child", "", "New managed object that will be added as a child addition (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("child", "managedObject.id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("managedObject"),
		flags.WithPowershellName("Add-ChildAddition"),
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
				return client.ManagedObjects.ChildAdditions.Assign(ctx, parentID, childID)
			})
		}, nil
	})
}
