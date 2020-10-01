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

type newApplicationCmd struct {
	*baseCmd
}

func newNewApplicationCmd() *newApplicationCmd {
	ccmd := &newApplicationCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new application",
		Long:  `Create a new application using explicit settings`,
		Example: `
$ c8y applications create --name myapp --type HOSTED --key "myapp-key" --contextPath "myapp"
Create a new hosted application
		`,
		RunE: ccmd.newApplication,
	}

	cmd.SilenceUsage = true

	addDataFlag(cmd)
	cmd.Flags().String("name", "", "Name of application (required)")
	cmd.Flags().String("key", "", "Shared secret of application (required)")
	cmd.Flags().String("type", "", "Type of application. Possible values are EXTERNAL, HOSTED, MICROSERVICE (required)")
	cmd.Flags().String("availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().String("contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().String("resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")
	cmd.Flags().String("resourcesUsername", "", "authorization username to access resourcesUrl")
	cmd.Flags().String("resourcesPassword", "", "authorization password to access resourcesUrl")
	cmd.Flags().String("externalUrl", "", "URL to the external application. Required when application type is EXTERNAL")

	// Required flags
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("type")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newApplicationCmd) newApplication(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") || globalUseNonDefaultPageSize {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}
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
	if v, err := cmd.Flags().GetString("name"); err == nil {
		if v != "" {
			body.Set("name", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "name", err))
	}
	if v, err := cmd.Flags().GetString("key"); err == nil {
		if v != "" {
			body.Set("key", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "key", err))
	}
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			body.Set("type", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	if v, err := cmd.Flags().GetString("availability"); err == nil {
		if v != "" {
			body.Set("availability", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "availability", err))
	}
	if v, err := cmd.Flags().GetString("contextPath"); err == nil {
		if v != "" {
			body.Set("contextPath", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "contextPath", err))
	}
	if v, err := cmd.Flags().GetString("resourcesUrl"); err == nil {
		if v != "" {
			body.Set("resourcesUrl", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "resourcesUrl", err))
	}
	if v, err := cmd.Flags().GetString("resourcesUsername"); err == nil {
		if v != "" {
			body.Set("resourcesUsername", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "resourcesUsername", err))
	}
	if v, err := cmd.Flags().GetString("resourcesPassword"); err == nil {
		if v != "" {
			body.Set("resourcesPassword", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "resourcesPassword", err))
	}
	if v, err := cmd.Flags().GetString("externalUrl"); err == nil {
		if v != "" {
			body.Set("externalUrl", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "externalUrl", err))
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/application/applications", pathParameters)

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

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
