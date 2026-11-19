// v2-based inventory assets get (deprecated; use `inventory children get
// --childType asset`): fetches a single child-asset reference from a parent
// (resolved as a managed object) and returns the referenced managed object. The
// id flag drives iteration.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/children/childref"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get child asset reference
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get child asset reference",
		Long:  `Get managed object child asset reference`,
		Example: heredoc.Doc(`
$ c8y inventory assets get --id 12345 --child 12345
Get an existing child asset reference
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Managed object id (required) (accepts pipeline)")
	cmd.Flags().StringSlice("child", []string{""}, "Child managed object id (required)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Get-ChildAsset"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReference+json", ""),
	)

	_ = cmd.MarkFlagRequired("child")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
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
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("id")), nil)
		if err != nil {
			return nil, err
		}
		childID := childref.First(in.StringSlice("child"))
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.ManagedObjects.ChildAssets.Get(ctx, parentID, childID))
		}, nil
	})
}
