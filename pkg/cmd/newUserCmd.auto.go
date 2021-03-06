// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// NewUserCmd command
type NewUserCmd struct {
	*baseCmd
}

// NewNewUserCmd creates a command to Create user
func NewNewUserCmd() *NewUserCmd {
	ccmd := &NewUserCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create user",
		Long:  `Create a new user so that they can access the tenant`,
		Example: `
$ c8y users create --userName "testuser1" --password "a0)8k2kld9lm,!"
Create a user
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("userName", "", "User name, unique for a given domain. Max: 1000 characters (accepts pipeline)")
	cmd.Flags().String("firstName", "", "User first name")
	cmd.Flags().String("lastName", "", "User last name")
	cmd.Flags().String("phone", "", "User phone number. Format: '+[country code][number]', has to be a valid MSISDN")
	cmd.Flags().String("email", "", "User email address")
	cmd.Flags().Bool("enabled", false, "User activation status (true/false)")
	cmd.Flags().String("password", "", "User password. Min: 6, max: 32 characters. Only Latin1 chars allowed")
	cmd.Flags().Bool("sendPasswordResetEmail", false, "User activation status (true/false)")
	cmd.Flags().String("customProperties", "", "Custom properties to be added to the user")
	addProcessingModeFlag(cmd)
	addTemplateFlag(cmd)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("sendPasswordResetEmail", "true", "false"),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("userName", "userName", false, "id"),
	)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *NewUserCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringValue("userName", "userName"),
		flags.WithStringValue("firstName", "firstName"),
		flags.WithStringValue("lastName", "lastName"),
		flags.WithStringValue("phone", "phone"),
		flags.WithStringValue("email", "email"),
		flags.WithBoolValue("enabled", "enabled", ""),
		flags.WithStringValue("password", "password"),
		flags.WithBoolValue("sendPasswordResetEmail", "sendPasswordResetEmail", ""),
		flags.WithDataValue("customProperties", "customProperties"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
		flags.WithRequiredProperties("userName"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("user/{tenant}/users")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
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
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
