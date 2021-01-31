// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type NewTenantCmd struct {
	*baseCmd
}

func NewNewTenantCmd() *NewTenantCmd {
	var _ = fmt.Errorf
	ccmd := &NewTenantCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "New tenant",
		Long:  ``,
		Example: `
$ c8y tenants create --company "mycompany" --domain "mycompany" --adminName "admin" --password "mys3curep9d8"
Create a new tenant (from the management tenant)
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("company", "", "Company name. Maximum 256 characters (required)")
	cmd.Flags().String("domain", "", "Domain name to be used for the tenant. Maximum 256 characters (required)")
	cmd.Flags().String("adminName", "", "Username of the tenant administrator")
	cmd.Flags().String("adminPass", "", "Password of the tenant administrator")
	cmd.Flags().String("contactName", "", "A contact name, for example an administrator, of the tenant")
	cmd.Flags().String("contactPhone", "", "An international contact phone number")
	cmd.Flags().String("tenantId", "", "The tenant ID. This should be left bank unless you know what you are doing. Will be auto-generated if not present.")
	addDataFlag(cmd)
	addProcessingModeFlag(cmd)

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	cmd.MarkFlagRequired("company")
	cmd.MarkFlagRequired("domain")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *NewTenantCmd) RunE(cmd *cobra.Command, args []string) error {
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err := flags.WithQueryOptions(
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

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()
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
	if v, err := cmd.Flags().GetString("tenantId"); err == nil {
		if v != "" {
			body.Set("tenantId", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "tenantId", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/tenant/tenants", pathParameters)

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

	return processRequestAndResponseWithWorkers(cmd, &req, "")
}
