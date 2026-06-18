// v2-based microservice list: fills the typed microservices.ListOptions
// (name/owner/providedFor/subscriber/user filters) and streams
// Microservices.ListAll. The user flag is the iterating input (pipe or
// --user), matching v1, so the command runs once for the common no-pipe case.
// The SDK always scopes the query to the MICROSERVICE application type, so the
// --type flag is kept only for surface parity.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/microservices"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get microservice collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get microservice collection",
		Long: `Get a collection of microservices in the current tenant
`,
		Example: heredoc.Doc(`
$ c8y microservices list --pageSize 100
Get microservices

$ c8y microservices list --name cockpit
Get a microservice by name

$ c8y microservices list --name device-simulator --user myuser
Check if a user has access to the device-simulator microservice

$ c8y microservices list --owner t12345
List all microservices owned by specific tenant
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "MICROSERVICE", "Application type")
	cmd.Flags().String("name", "", "The name of the application.")
	cmd.Flags().String("owner", "", "The ID of the tenant that owns the applications.")
	cmd.Flags().String("providedFor", "", "The ID of a tenant that is subscribed to the applications but doesn't own them.")
	cmd.Flags().String("subscriber", "", "The ID of a tenant that is subscribed to the applications.")
	cmd.Flags().StringSlice("user", []string{""}, "The ID of a user that has access to the applications. (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "MICROSERVICE"),
		completion.WithApplication("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("owner", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("providedFor", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("subscriber", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("user", "user", false, "id"),
		flags.WithPipelineAliases("name", "id"),
		flags.WithPipelineAliases("owner", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("providedFor", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("subscriber", "tenant", "owner.tenant.id"),
		flags.WithCollectionProperty("applications"),
		flags.WithPowershellName("Get-MicroserviceCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationCollection+json", "application/vnd.com.nsn.cumulocity.application+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("user"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	rawOutput := r.Config.RawOutput()
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := microservices.ListOptions{
			Name:        in.String("name"),
			Owner:       in.String("owner"),
			ProvidedFor: in.String("providedFor"),
			Subscriber:  in.String("subscriber"),
			User:        in.String("user"),
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Microservices.ListAll), nil
	})
}
