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

// DisableMicroserviceCmd command
type DisableMicroserviceCmd struct {
	*subcommand.SubCommand
}

// NewDisableMicroserviceCmd creates a command to unsubscribe microservice
func NewDisableMicroserviceCmd() *DisableMicroserviceCmd {
	ccmd := &DisableMicroserviceCmd{}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "unsubscribe microservice",
		Long: `Disable (unsubscribe) a microservice from the current tenant
`,
		Example: heredoc.Doc(`
$ c8y microservices disable --id 12345
Disable (unsubscribe) to a microservice

$ c8y microservices disable --id myapp
Disable (unsubscribe) to a microservice
        `),
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant id")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DisableMicroserviceCmd) RunE(cmd *cobra.Command, args []string) error {
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
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/tenant/tenants/{tenant}/applications/{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		WithMicroserviceByNameFirstMatch(args, "id", "id"),
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "DELETE",
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
