// Code generated from specification version 1.0.0: DO NOT EDIT
package addroletouser

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// AddRoleToUserCmd command
type AddRoleToUserCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	Config  func() (*config.Config, error)
	Client  func() (*c8y.Client, error)
}

// NewAddRoleToUserCmd creates a command to Add Role to user
func NewAddRoleToUserCmd(f *cmdutil.Factory) *AddRoleToUserCmd {
	ccmd := &AddRoleToUserCmd{
		factory: f,
		Config:  f.Config,
		Client:  f.Client,
	}
	cmd := &cobra.Command{
		Use:   "addRoleToUser",
		Short: "Add Role to user",
		Long:  `Add a role to an existing user`,
		Example: heredoc.Doc(`
$ c8y userRoles addRoleToUser --user "myuser" --role "ROLE_ALARM_READ"
Add a role (ROLE_ALARM_READ) to a user
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("user", []string{""}, "User prefix or full username (required)")
	cmd.Flags().StringSlice("role", []string{""}, "User role id (accepts pipeline)")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("role", "role.self", false, "self", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("user")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AddRoleToUserCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.Config()
	if err != nil {
		return err
	}
	client, err := n.Client()
	if err != nil {
		return err
	}
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
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
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/user/{tenant}/users/{user}/roles")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
		c8yfetcher.WithUserByNameFirstMatch(client, args, "user", "user"),
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
		DryRun:       cfg.DryRun(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
