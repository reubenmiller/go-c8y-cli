// v2-based application list: fills the typed applications.ListOptions (name/
// owner/type/availability/... filters) and streams Applications.ListAll. The type
// flag is the iterating input (pipe or --type), so the command runs once for the
// common no-pipe case.
package list

import (
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get application collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get application collection",
		Long:  `Get a collection of applications by a given filter`,
		Example: heredoc.Doc(`
$ c8y applications list --pageSize 100
Get applications

$ c8y applications list --type HOSTED
Get hosted applications
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "Application type (accepts pipeline)")
	cmd.Flags().String("name", "", "The name of the application.")
	cmd.Flags().String("owner", "", "The ID of the tenant that owns the applications.")
	cmd.Flags().String("providedFor", "", "The ID of a tenant that is subscribed to the applications but doesn't own them.")
	cmd.Flags().String("subscriber", "", "The ID of a tenant that is subscribed to the applications.")
	cmd.Flags().StringSlice("user", []string{""}, "The ID of a user that has access to the applications.")
	cmd.Flags().String("tenant", "", "The ID of a tenant that either owns the application or is subscribed to the applications.")
	cmd.Flags().Bool("hasVersions", false, "When set to true, the returned result contains applications with an applicationVersions field that is not empty. When set to false, the result will contain applications with an empty applicationVersions field.")
	cmd.Flags().String("availability", "", "Application access level for other tenants.")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "APAMA_CEP_RULE", "EXTERNAL", "HOSTED", "MICROSERVICE"),
		completion.WithApplication("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("owner", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("providedFor", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("subscriber", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("availability", "SHARED", "PRIVATE", "MARKET"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("type", "type", false, "id"),
		flags.WithCollectionProperty("applications"),
		flags.WithPowershellName("Get-ApplicationCollection"),
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

	if err := r.InputFlag("type"); err != nil {
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
	hasVersionsChanged := cmd.Flags().Changed("hasVersions")

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := applications.ListOptions{
			Type:         in.String("type"),
			Name:         in.String("name"),
			Owner:        in.String("owner"),
			ProvidedFor:  in.String("providedFor"),
			Subscriber:   in.String("subscriber"),
			User:         strings.Join(in.StringSlice("user"), ","),
			Tenant:       in.String("tenant"),
			Availability: in.String("availability"),
		}
		if hasVersionsChanged {
			opt.HasVersions = strconv.FormatBool(in.Bool("hasVersions"))
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Applications.ListAll), nil
	})
}
