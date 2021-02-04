// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
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
	var err error
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = flags.WithQueryParameters(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}
	err = flags.WithQueryOptions(
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

	err = flags.WithHeaders(
		cmd,
		headers,
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return newUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
	)
	if err != nil {
		return newUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		flags.WithDataValue(FlagDataName),
		flags.WithStringValue("company", "company"),
		flags.WithStringValue("domain", "domain"),
		flags.WithStringValue("adminName", "adminName"),
		flags.WithStringValue("adminPass", "adminPass"),
		flags.WithStringValue("contactName", "contactName"),
		flags.WithStringValue("contactPhone", "contact_phone"),
		flags.WithStringValue("tenantId", "tenantId"),
	)
	if err != nil {
		return newUserError(err)
	}

	if err := setLazyDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	err = flags.WithPathParameters(
		cmd,
		pathParameters,
	)

	path := replacePathParameters("/tenant/tenants", pathParameters)

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path,
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, PipeOption{"", false})
}
