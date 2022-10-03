// Code generated from specification version 1.0.0: DO NOT EDIT
package find

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// FindCmd command
type FindCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewFindCmd creates a command to Find services
func NewFindCmd(f *cmdutil.Factory) *FindCmd {
	ccmd := &FindCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find services",
		Long:  `Find services of any device`,
		Example: heredoc.Doc(`
$ c8y devices services find
Find all services (from any device)

$ c8y devices services find --status down
Find any services which are currently down

$ c8y devices services find --name ntp --status down
Find any ntp services which are currently down
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().String("serviceType", "", "Filter by service type e.g. systemd")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("status", "", "Filter by service status (custom values allowed)")
	cmd.Flags().String("orderBy", "", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "up", "down", "unknown"),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithCollectionProperty("managedObjects"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *FindCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
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
		flags.WithBoolValue("withParents", "withParents", ""),

		flags.WithCumulocityQuery(
			[]flags.GetOption{
				flags.WithStaticStringValue("type", "(type eq 'c8y_Service')"),
				flags.WithStringValue("query", "query", "%s"),
				flags.WithStringValue("serviceType", "serviceType", "(serviceType eq '%s')"),
				flags.WithStringValue("name", "name", "(name eq '%s')"),
				flags.WithStringValue("status", "status", "(status eq '%s')"),
			},
			"query",
		),
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
	path := flags.NewStringTemplate("inventory/managedObjects")
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
