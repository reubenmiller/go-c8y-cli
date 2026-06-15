// v2-based agent get: the id flag drives iteration (pipe or --id); each agent
// reference (id or name, resolved scoped to agents) is fetched via Agents.Get.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/agents"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get agent
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get agent",
		Long:  `Get an agent's managed object representation`,
		Example: heredoc.Doc(`
$ c8y agents get --id 12345
Get agent by id

$ c8y agents get --id myAgent
Get agent by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Agent ID (required) (accepts pipeline)")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")
	cmd.Flags().Bool("withLatestValues", false, "(FEATURE_PREVIEW) Include c8y_LatestMeasurements fragment, which contains the latest measurement values reported by the device to the platform")

	completion.WithOptions(
		cmd,
		completion.WithAgent("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Get-Agent"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.customAgent+json", ""),
	)

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
		ref := agents.AgentRef(c8ystream.NameOrID(in.String("id")))
		// GAP: --withGroups has no field in managedobjects.GetOptions yet.
		opt := agents.GetOptions{
			SkipChildrenNames: in.Bool("skipChildrenNames"),
			WithChildren:      in.Bool("withChildren"),
			WithChildrenCount: in.Bool("withChildrenCount"),
			WithParents:       in.Bool("withParents"),
			WithLatestValues:  in.Bool("withLatestValues"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Agents.Get(ctx, ref, opt))
		}, nil
	})
}
