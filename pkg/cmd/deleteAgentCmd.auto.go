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

type deleteAgentCmd struct {
	*baseCmd
}

func newDeleteAgentCmd() *deleteAgentCmd {
	ccmd := &deleteAgentCmd{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete agent",
		Long: `Delete an agent from the platform. This will delete all data associated to the agent
(i.e. alarms, events, operations and measurements)
`,
		Example: `
$ c8y agents delete --id 12345
Get agent by id
		`,
		RunE: ccmd.deleteAgent,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Agent ID (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteAgentCmd) deleteAgent(cmd *cobra.Command, args []string) error {

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

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("id") {
		idInputValues, idValue, err := getFormattedAgentSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching agents found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching agents found", idInputValues)
		}

		for _, item := range idValue {
			if item != "" {
				pathParameters["id"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{id}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "DELETE",
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
