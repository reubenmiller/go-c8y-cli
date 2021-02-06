// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type AddManagedObjectChildAdditionCmd struct {
	*baseCmd
}

func NewAddManagedObjectChildAdditionCmd() *AddManagedObjectChildAdditionCmd {
	ccmd := &AddManagedObjectChildAdditionCmd{}
	cmd := &cobra.Command{
		Use:   "createChildAddition",
		Short: "Add a managed object as a child addition to another existing managed object",
		Long:  `Add a managed object as a child addition to another existing managed object`,
		Example: `
$ c8y inventoryReferences createChildAddition --id 12345 --newChild 6789
Add a related managed object as a child to an existing managed object
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Managed object id where the child addition will be added to (required)")
	cmd.Flags().String("newChild", "", "New managed object that will be added as a child addition (required) (accepts pipeline)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("newChild"),
	)

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *AddManagedObjectChildAdditionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	query := url.Values{}
	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		WithDataValue(),
		flags.WithStringValue("newChild", "managedObject.id"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
		flags.WithStringValue("id", "id"),
	)
	if err != nil {
		return err
	}

	path := replacePathParameters("inventory/managedObjects/{id}/childAdditions", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	pipeOption := PipeOption{
		Name:              "newChild",
		Property:          "managedObject.id",
		Required:          true,
		ResolveByNameType: "",
		IteratorType:      "body",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
