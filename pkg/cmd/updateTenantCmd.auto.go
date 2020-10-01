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
		RunE: ccmd.updateTenant,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Tenant id")
	cmd.Flags().String("company", "", "Company name. Maximum 256 characters")
	cmd.Flags().String("domain", "", "Domain name to be used for the tenant. Maximum 256 characters (required)")
	cmd.Flags().String("adminName", "", "Username of the tenant administrator")
	cmd.Flags().String("adminPass", "", "Password of the tenant administrator")
	cmd.Flags().String("contactName", "", "A contact name, for example an administrator, of the tenant")
	cmd.Flags().String("contactPhone", "", "An international contact phone number")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("domain")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateTenantCmd) updateTenant(cmd *cobra.Command, args []string) error {

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
