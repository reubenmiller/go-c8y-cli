// v2-based devices children get: fetches a single child reference of the
// selected --childType (addition|asset|device) from a parent (resolved as a
// managed object) and returns the referenced managed object. The id flag drives
// iteration.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/children/childref"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get child
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get child",
		Long:  `Get a child of a device`,
		Example: heredoc.Doc(`
$ c8y devices children get --id 12345 --child 12345 --childType addition
Get an existing child addition reference
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Managed object id (required) (accepts pipeline)")
	cmd.Flags().String("childType", "", "Child relationship type (required)")
	cmd.Flags().StringSlice("child", []string{""}, "Child managed object id (required)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("childType", "addition", "asset", "device"),
		completion.WithDevice("child", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("id", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("child", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("managedObject"),
		flags.WithPowershellName("Get-DeviceChild"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObject+json", ""),
	)

	_ = cmd.MarkFlagRequired("childType")
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
		svc, err := childref.For(client, in.String("childType"))
		if err != nil {
			return nil, err
		}
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("id")), nil)
		if err != nil {
			return nil, err
		}
		childID := childref.First(in.StringSlice("child"))
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(svc.Get(ctx, parentID, childID))
		}, nil
	})
}
