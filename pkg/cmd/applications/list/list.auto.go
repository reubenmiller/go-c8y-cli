// Code generated from specification version 1.0.0: DO NOT EDIT
package list

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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

$ c8y applications list --name cockpit
Get an application by name

$ c8y applications list --name cockpit --user myuser
Check if a user has access to the cockpit application
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

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "APAMA_CEP_RULE", "EXTERNAL", "HOSTED", "MICROSERVICE"),
		completion.WithApplication("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("owner", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("providedFor", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("subscriber", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("type", "type", false, "id"),
		flags.WithPipelineAliases("name", "id"),
		flags.WithPipelineAliases("owner", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("providedFor", "tenant", "owner.tenant.id"),
		flags.WithPipelineAliases("subscriber", "tenant", "owner.tenant.id"),

		flags.WithCollectionProperty("applications"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("owner", "owner"),
		flags.WithStringValue("providedFor", "providedFor"),
		flags.WithStringValue("subscriber", "subscriber"),
		c8yfetcher.WithUserByNameFirstMatch(n.factory, args, "user", "user"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(false)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/application/applications")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
