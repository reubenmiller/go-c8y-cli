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

type newTenantOptionCmd struct {
	*baseCmd
}

func newNewTenantOptionCmd() *newTenantOptionCmd {
	ccmd := &newTenantOptionCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "New tenant option",
		Long:  ``,
		Example: `
$ c8y tenantOptions create --category "c8y_cli_tests" --key "option1" --value "1"
Create a tenant option
		`,
		RunE: ccmd.newTenantOption,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Category of option (required)")
	cmd.Flags().String("key", "", "Key of option (required)")
	cmd.Flags().String("value", "", "Value of option (required)")

	// Required flags
	cmd.MarkFlagRequired("category")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newTenantOptionCmd) newTenantOption(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("category"); err == nil {
		if v != "" {
			body.Set("category", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "category", err))
	}
	if v, err := cmd.Flags().GetString("key"); err == nil {
		if v != "" {
			body.Set("key", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "key", err))
	}
	if v, err := cmd.Flags().GetString("value"); err == nil {
		if v != "" {
			body.Set("value", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "value", err))
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/tenant/options", pathParameters)

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
