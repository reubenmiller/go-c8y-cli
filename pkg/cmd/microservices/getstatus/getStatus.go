// v2-based microservice status get: the id flag drives iteration (pipe or
// --id); the microservice reference is resolved to an id, then the status
// managed object(s) (inventory type c8y_Application_<id>) are listed via
// Microservices.GetStatus.
package getstatus

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetStatusCmd command
type GetStatusCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetStatusCmd creates a command to Get microservice status
func NewGetStatusCmd(f *cmdutil.Factory) *GetStatusCmd {
	ccmd := &GetStatusCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getStatus",
		Short: "Get microservice status",
		Long: `Get the status of a microservice which is stored as a managed object
`,
		Example: heredoc.Doc(`
$ c8y microservices getStatus --id 1234
Get microservice status

$ c8y microservices list | c8y microservices getStatus
Get microservice status (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "application.id", "id"),
		flags.WithPipelineAliases("id", "id"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-MicroserviceStatus"),
		flags.WithOutputType("application/json", "application/vnd.com.nsn.cumulocity.application+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetStatusCmd) RunE(cmd *cobra.Command, args []string) error {
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
		// The status query embeds the resolved id in the type filter
		// (c8y_Application_<id>), so resolve the reference first.
		id, err := client.Microservices.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("id")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.Microservices.GetStatus(ctx, id).Items())
		}, nil
	})
}
