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

type updateTenantOptionEditableCmd struct {
	*baseCmd
}

func newUpdateTenantOptionEditableCmd() *updateTenantOptionEditableCmd {
	ccmd := &updateTenantOptionEditableCmd{}

	cmd := &cobra.Command{
		Use:   "updateEdit",
		Short: "Update tenant option editibility",
		Long: `Required role:: ROLE_OPTION_MANAGEMENT_ADMIN, Required tenant management Example Request:: Update access.control.allow.origin option.
`,
		Example: `
$ c8y tenantOptions updateEdit --category "c8y_cli_tests" --key "option8" --editable "true"
Update editable property for an existing tenant option
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.updateTenantOptionEditable,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Tenant Option category (required)")
	cmd.Flags().String("key", "", "Tenant Option key (required)")
	cmd.Flags().String("editable", "", "Whether the tenant option should be editable or not (required)")

	// Required flags
	cmd.MarkFlagRequired("category")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("editable")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateTenantOptionEditableCmd) updateTenantOptionEditable(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("editable"); err == nil {
		if v != "" {
			body.Set("editable", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "editable", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

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

	path := replacePathParameters("/tenant/options/{category}/{key}/editable", pathParameters)

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
