// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// UpdateUserCurrentCmd command
type UpdateUserCurrentCmd struct {
	*subcommand.SubCommand
}

// NewUpdateUserCurrentCmd creates a command to Update current user
func NewUpdateUserCurrentCmd() *UpdateUserCurrentCmd {
	ccmd := &UpdateUserCurrentCmd{}
	cmd := &cobra.Command{
		Use:   "updateCurrentUser",
		Short: "Update current user",
		Long: `Update properties or settings of your user such as first/last name, email or password
`,
		Example: heredoc.Doc(`
$ c8y users updateCurrentUser --lastName "Smith"
Update the current user's last name
        `),
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("firstName", "", "User first name")
	cmd.Flags().String("lastName", "", "User last name")
	cmd.Flags().String("phone", "", "User phone number. Format: '+[country code][number]', has to be a valid MSISDN")
	cmd.Flags().String("email", "", "User email address")
	cmd.Flags().String("enabled", "", "User activation status (true/false)")
	cmd.Flags().String("password", "", "User password. Min: 6, max: 32 characters. Only Latin1 chars allowed")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("enabled", "true", "false"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateUserCurrentCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
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
		WithDataValue(),
		flags.WithStringValue("firstName", "firstName"),
		flags.WithStringValue("lastName", "lastName"),
		flags.WithStringValue("phone", "phone"),
		flags.WithStringValue("email", "email"),
		flags.WithStringValue("enabled", "enabled"),
		flags.WithStringValue("password", "password"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/user/currentUser")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
		DryRun:       cliConfig.DryRun(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
