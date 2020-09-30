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

type getSystemOptionCmd struct {
	*baseCmd
}

func newGetSystemOptionCmd() *getSystemOptionCmd {
	ccmd := &getSystemOptionCmd{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get system option",
		Long:  `Get a system option by category and key`,
		Example: `
$ c8y systemOptions get --category "system" --key "version"
Get a list of system options
		`,
		RunE: ccmd.getSystemOption,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "System Option category (required)")
	cmd.Flags().String("key", "", "System Option key (required)")

	// Required flags
	cmd.MarkFlagRequired("category")
	cmd.MarkFlagRequired("key")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getSystemOptionCmd) getSystemOption(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("pageSize") {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}

	if cmd.Flags().Changed("withTotalPages") {
		if v, err := cmd.Flags().GetBool("withTotalPages"); err == nil && v {
			query.Add("withTotalPages", "true")
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

	// path parameters
	pathParameters := make(map[string]string)
	if v, err := cmd.Flags().GetString("category"); err == nil {
		if v != "" {
			pathParameters["category"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "category", err))
	}
	if v, err := cmd.Flags().GetString("key"); err == nil {
		if v != "" {
			pathParameters["key"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "key", err))
	}

	path := replacePathParameters("/tenant/system/options/{category}/{key}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "GET",
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
