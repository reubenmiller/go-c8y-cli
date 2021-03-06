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

type updateTenantCmd struct {
	*baseCmd
}

func newUpdateTenantCmd() *updateTenantCmd {
	ccmd := &updateTenantCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update tenant",
		Long:  ``,
		Example: `
$ c8y tenants update --id "mycompany" --contactName "John Smith"
Update a tenant by name (from the mangement tenant)
        `,
		PreRunE: validateUpdateMode,
		RunE:    ccmd.updateTenant,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Tenant id")
	cmd.Flags().String("company", "", "Company name. Maximum 256 characters")
	cmd.Flags().String("domain", "", "Domain name to be used for the tenant. Maximum 256 characters")
	cmd.Flags().String("adminName", "", "Username of the tenant administrator")
	cmd.Flags().String("adminPass", "", "Password of the tenant administrator")
	cmd.Flags().String("contactName", "", "A contact name, for example an administrator, of the tenant")
	cmd.Flags().String("contactPhone", "", "An international contact phone number")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateTenantCmd) updateTenant(cmd *cobra.Command, args []string) error {

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
	if cmd.Flags().Changed("processingMode") {
		if v, err := cmd.Flags().GetString("processingMode"); err == nil && v != "" {
			headers.Add("X-Cumulocity-Processing-Mode", v)
		}
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if v, err := cmd.Flags().GetString("company"); err == nil {
		if v != "" {
			body.Set("company", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "company", err))
	}
	if v, err := cmd.Flags().GetString("domain"); err == nil {
		if v != "" {
			body.Set("domain", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "domain", err))
	}
	if v, err := cmd.Flags().GetString("adminName"); err == nil {
		if v != "" {
			body.Set("adminName", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "adminName", err))
	}
	if v, err := cmd.Flags().GetString("adminPass"); err == nil {
		if v != "" {
			body.Set("adminPass", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "adminPass", err))
	}
	if v, err := cmd.Flags().GetString("contactName"); err == nil {
		if v != "" {
			body.Set("contactName", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "contactName", err))
	}
	if v, err := cmd.Flags().GetString("contactPhone"); err == nil {
		if v != "" {
			body.Set("contact_phone", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "contactPhone", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	if v := getTenantWithDefaultFlag(cmd, "id", client.TenantName); v != "" {
		pathParameters["id"] = v
	}

	path := replacePathParameters("/tenant/tenants/{id}", pathParameters)

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
