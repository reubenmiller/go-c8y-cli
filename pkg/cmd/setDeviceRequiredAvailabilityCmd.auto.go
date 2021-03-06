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

// SetDeviceRequiredAvailabilityCmd command
type SetDeviceRequiredAvailabilityCmd struct {
	*baseCmd
}

// NewSetDeviceRequiredAvailabilityCmd creates a command to Set required availability
func NewSetDeviceRequiredAvailabilityCmd() *SetDeviceRequiredAvailabilityCmd {
	ccmd := &SetDeviceRequiredAvailabilityCmd{}
	cmd := &cobra.Command{
		Use:   "setRequiredAvailability",
		Short: "Set required availability",
		Long:  `Set the required availability of a device. Devices that have not sent any message in the response interval are considered unavailable. Response interval can have value between -32768 and 32767 and any values out of range will be shrink to range borders. Such devices are marked as unavailable (see below) and an unavailability alarm is raised. Devices with a response interval of zero minutes are considered to be under maintenance. No alarm is raised while a device is under maintenance. Devices that do not contain 'c8y_RequiredAvailability' are not monitored.`,
		Example: `
$ c8y inventory setRequiredAvailability --device 12345 --interval 10
Set the required availability of a device by name to 10 minutes
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID (required) (accepts pipeline)")
	cmd.Flags().Int("interval", 0, "Interval in minutes (required)")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("interval")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *SetDeviceRequiredAvailabilityCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithIntValue("interval", "c8y_RequiredAvailability.responseInterval"),
		WithTemplateValue(),
		WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{device}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		WithDeviceByNameFirstMatch(args, "device", "device"),
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
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
