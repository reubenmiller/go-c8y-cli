// Code generated from specification version 1.0.0: DO NOT EDIT
package addusertogroup

import (
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

// AddUserToGroupCmd command
type AddUserToGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAddUserToGroupCmd creates a command to Add user to group
func NewAddUserToGroupCmd(f *cmdutil.Factory) *AddUserToGroupCmd {
	ccmd := &AddUserToGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "addUserToGroup",
		Short: "Add user to group",
		Long:  `Add an existing user to a group`,
		Example: heredoc.Doc(`
$ c8y userreferences addUserToGroup --group 1 --user peterpi@example.com
Add a user to a user group

$ c8y users list | c8y userreferences addUserToGroup --group admins
Add a list of users to admins group (using pipeline)

$ c8y users list | c8y userreferences addUserToGroup --group business | c8y userreferences addUserToGroup --group admins
Add a list of users to business and admins group (using pipeline)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group ID (required)")
	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("user", []string{""}, "User id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithUserGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("user", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("user", "user.self", true, "user.id", "id", "self"),
		flags.WithPipelineAliases("group", "id"),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AddUserToGroupCmd) RunE(cmd *cobra.Command, args []string) error {
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
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

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
		flags.WithProcessingModeValue(),
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
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		c8yfetcher.WithUserSelfByNameFirstMatch(n.factory, args, "user", "user.self"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("user.self"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/user/{tenant}/groups/{group}/users")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithUserGroupByNameFirstMatch(n.factory, args, "group", "group"),
		flags.WithStringDefaultValue(n.factory.GetTenant(), "tenant", "tenant"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "POST",
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
