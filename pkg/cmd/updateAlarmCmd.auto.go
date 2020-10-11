// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type updateAlarmCmd struct {
	*baseCmd
}

func newUpdateAlarmCmd() *updateAlarmCmd {
	ccmd := &updateAlarmCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an alarm",
		Long:  `Update an alarm by its id`,
		Example: `
$ c8y alarms update --id 12345 --status ACKNOWLEDGED
Acknowledge an existing alarm

$ c8y alarms update --id 12345 --severity CRITICAL
Update severity of an existing alarm to CRITICAL
		`,
		RunE: ccmd.updateAlarm,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Alarm id (required)")
	cmd.Flags().String("status", "", "Comma separated alarm statuses, for example ACTIVE,CLEARED.")
	cmd.Flags().String("severity", "", "Alarm severity, for example CRITICAL, MAJOR, MINOR or WARNING.")
	cmd.Flags().String("text", "", "Text description of the alarm.")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateAlarmCmd) updateAlarm(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetString("status"); err == nil {
		if v != "" {
			body.Set("status", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "status", err))
	}
	if v, err := cmd.Flags().GetString("severity"); err == nil {
		if v != "" {
			body.Set("severity", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "severity", err))
	}
	if v, err := cmd.Flags().GetString("text"); err == nil {
		if v != "" {
			body.Set("text", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "text", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("alarm/alarms/{id}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "PUT",
		Path:         path,
		Query:        queryValue,
		Body:         body.GetMap(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
