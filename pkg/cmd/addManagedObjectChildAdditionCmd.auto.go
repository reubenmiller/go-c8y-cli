// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
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
	var _ = fmt.Errorf
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

	cmd.Flags().String("id", "", "Managed object id where the child addition will be added to (required) (accepts pipeline)")
	cmd.Flags().StringSlice("newChild", []string{""}, "New managed object that will be added as a child addition (required)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("id"),
	)

	// Required flags
	cmd.MarkFlagRequired("newChild")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *AddManagedObjectChildAdditionCmd) RunE(cmd *cobra.Command, args []string) error {
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err := flags.WithQueryOptions(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}

	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if items, err := cmd.Flags().GetStringSlice("newChild"); err == nil {
		if len(items) > 0 {
			for _, v := range items {
				if v != "" {
					body.Set("managedObject.id", v)
				}
			}
		}
	} else {
		return newUserError("Flag does not exist")
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("inventory/managedObjects/{id}/childAdditions", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path,
		Query:        queryValue,
		Body:         body.GetMap(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, "id")
}
