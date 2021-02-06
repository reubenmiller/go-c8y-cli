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

type GetManagedObjectChildDeviceReferenceCmd struct {
	*baseCmd
}

func NewGetManagedObjectChildDeviceReferenceCmd() *GetManagedObjectChildDeviceReferenceCmd {
	ccmd := &GetManagedObjectChildDeviceReferenceCmd{}
	cmd := &cobra.Command{
		Use:   "getChildDevice",
		Short: "Get managed object child device reference",
		Long:  ``,
		Example: `
$ c8y inventoryReferences getChildDevice --device 12345 --reference 12345
Get an existing child device reference
        `,
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "ManagedObject id (required) (accepts pipeline)")
	cmd.Flags().StringSlice("reference", []string{""}, "Device reference id (required)")

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("device"),
	)

	// Required flags
	cmd.MarkFlagRequired("reference")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *GetManagedObjectChildDeviceReferenceCmd) RunE(cmd *cobra.Command, args []string) error {
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
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(&query)

	queryValue, err := url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
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
		WithDeviceByNameFirstMatch(args, "reference", "reference"),
	)
	if err != nil {
		return err
	}

	path := replacePathParameters("inventory/managedObjects/{device}/childDevices/{reference}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	pipeOption := PipeOption{
		Name:              "device",
		Property:          "",
		Required:          true,
		ResolveByNameType: "device",
		IteratorType:      "path",
	}
	return processRequestAndResponseWithWorkers(cmd, &req, pipeOption)
}
