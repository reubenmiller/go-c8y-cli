// v2-based microservice bootstrap-user get: the id flag drives iteration (pipe
// or --id); the microservice reference is resolved to an id, then the bootstrap
// user is fetched via Microservices.BootstrapUser.Get.
package getbootstrapuser

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

// GetBootstrapUserCmd command
type GetBootstrapUserCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetBootstrapUserCmd creates a command to Get microservice bootstrap user
func NewGetBootstrapUserCmd(f *cmdutil.Factory) *GetBootstrapUserCmd {
	ccmd := &GetBootstrapUserCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "getBootstrapUser",
		Short: "Get microservice bootstrap user",
		Long: `Get the bootstrap user associated to a microservice. The bootstrap user is required when running
a microservice locally (i.e. during development)
`,
		Example: heredoc.Doc(`
$ c8y microservices getBootstrapUser --id 12345
Get application bootstrap user by app id

$ c8y microservices getBootstrapUser --id report-agent
Get application bootstrap user by app name
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
		flags.WithPowershellName("Get-MicroserviceBootstrapUser"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.bootstrapuser+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetBootstrapUserCmd) RunE(cmd *cobra.Command, args []string) error {
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
		// The bootstrapUser sub-resource takes a plain id, so resolve the
		// reference (id or name) first (under ResolveContext so it works --dry).
		id, err := client.Microservices.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("id")), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Microservices.BootstrapUser.Get(ctx, id))
		}, nil
	})
}
