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

type NewManagedObjectChildDeviceCmd struct {
	*baseCmd
}

func NewNewManagedObjectChildDeviceCmd() *NewManagedObjectChildDeviceCmd {
	ccmd := &NewManagedObjectChildDeviceCmd{}
	cmd := &cobra.Command{
		Use:   "assignChildDevice",
		Short: "Create a child device reference",
		Long:  `Create a child device reference`,
		Example: `
$ c8y inventoryReferences assignChildDevice --device 12345 --newChild 44235
Assign a device as a child device to an existing device
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device. (required)")
	cmd.Flags().StringSlice("newChild", []string{""}, "New child device (required) (accepts pipeline)")
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("newChild"),
	)

	// Required flags
	cmd.MarkFlagRequired("device")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewManagedObjectChildDeviceCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}
	err = flags.WithQueryOptions(
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

	err = flags.WithHeaders(
		cmd,
		headers,
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		flags.WithDataValue(FlagDataName),
	)
	if err != nil {
		return newUserError(err)
	}

	if cmd.Flags().Changed("newChild") {
		newChildInputValues, newChildValue, err := getFormattedDeviceSlice(cmd, args, "newChild")

		if err != nil {
			return newUserError("no matching devices found", newChildInputValues, err)
		}

		if len(newChildValue) == 0 {
			return newUserError("no matching devices found", newChildInputValues)
		}

		for _, item := range newChildValue {
			if item != "" {
				body.Set("managedObject.id", newIDValue(item).GetID())
			}
		}
	}
	if err := setLazyDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)
	if cmd.Flags().Changed("device") {
		deviceInputValues, deviceValue, err := getFormattedDeviceSlice(cmd, args, "device")

		if err != nil {
			return newUserError("no matching devices found", deviceInputValues, err)
		}

		if len(deviceValue) == 0 {
			return newUserError("no matching devices found", deviceInputValues)
		}

		for _, item := range deviceValue {
			if item != "" {
				pathParameters["device"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{device}/childDevices", pathParameters)

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

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"newChild", true})
}
