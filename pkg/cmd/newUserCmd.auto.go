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

type newUserCmd struct {
	*baseCmd
}

func newNewUserCmd() *newUserCmd {
	ccmd := &newUserCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user within the collection",
		Long:  `This command can be used to grant a new user to the tenant`,
		Example: `
$ c8y users create --userName "testuser1" --password "a0)8k2kld9lm,!"
Create a user
        `,
		PreRunE: validateCreateMode,
		RunE:    ccmd.newUser,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("userName", "", "User name, unique for a given domain. Max: 1000 characters (required)")
	cmd.Flags().String("firstName", "", "User first name")
	cmd.Flags().String("lastName", "", "User last name")
	cmd.Flags().String("phone", "", "User phone number. Format: '+[country code][number]', has to be a valid MSISDN")
	cmd.Flags().String("email", "", "User email address")
	cmd.Flags().Bool("enabled", false, "User activation status (true/false)")
	cmd.Flags().String("password", "", "User password. Min: 6, max: 32 characters. Only Latin1 chars allowed")
	cmd.Flags().Bool("sendPasswordResetEmail", false, "User activation status (true/false)")
	cmd.Flags().String("customProperties", "", "Custom properties to be added to the user")

	// Required flags
	cmd.MarkFlagRequired("userName")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newUserCmd) newUser(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("userName"); err == nil {
		if v != "" {
			body.Set("userName", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "userName", err))
	}
	if v, err := cmd.Flags().GetString("firstName"); err == nil {
		if v != "" {
			body.Set("firstName", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "firstName", err))
	}
	if v, err := cmd.Flags().GetString("lastName"); err == nil {
		if v != "" {
			body.Set("lastName", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "lastName", err))
	}
	if v, err := cmd.Flags().GetString("phone"); err == nil {
		if v != "" {
			body.Set("phone", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "phone", err))
	}
	if v, err := cmd.Flags().GetString("email"); err == nil {
		if v != "" {
			body.Set("email", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "email", err))
	}
	if cmd.Flags().Changed("enabled") {
		if v, err := cmd.Flags().GetBool("enabled"); err == nil {
			body.Set("enabled", v)
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if v, err := cmd.Flags().GetString("password"); err == nil {
		if v != "" {
			body.Set("password", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "password", err))
	}
	if cmd.Flags().Changed("sendPasswordResetEmail") {
		if v, err := cmd.Flags().GetBool("sendPasswordResetEmail"); err == nil {
			body.Set("sendPasswordResetEmail", v)
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if cmd.Flags().Changed("customProperties") {
		if v, err := cmd.Flags().GetString("customProperties"); err == nil {
			body.Set("customProperties", MustParseJSON(v))
		} else {
			return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "customProperties", err))
		}
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)
	if v := getTenantWithDefaultFlag(cmd, "tenant", client.TenantName); v != "" {
		pathParameters["tenant"] = v
	}

	path := replacePathParameters("user/{tenant}/users", pathParameters)

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
