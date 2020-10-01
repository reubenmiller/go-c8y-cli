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

type newRetentionRuleCmd struct {
	*baseCmd
}

func newNewRetentionRuleCmd() *newRetentionRuleCmd {
	ccmd := &newRetentionRuleCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "New retention rule",
		Long: `Create a new retention rule to managed when data is deleted in the tenant
`,
		Example: `
$ c8y retentionRules create --dataType ALARM --maximumAge 180
Create a retention rule
		`,
		RunE: ccmd.newRetentionRule,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("dataType", "", "RetentionRule will be applied to this type of documents, possible values [ALARM, AUDIT, EVENT, MEASUREMENT, OPERATION, *]. (required)")
	cmd.Flags().String("fragmentType", "", "RetentionRule will be applied to documents with fragmentType.")
	cmd.Flags().String("type", "", "RetentionRule will be applied to documents with type.")
	cmd.Flags().String("source", "", "RetentionRule will be applied to documents with source.")
	cmd.Flags().Int("maximumAge", 0, "Maximum age of document in days. (required)")
	cmd.Flags().Bool("editable", false, "Whether the rule is editable. Can be updated only by management tenant.")

	// Required flags
	cmd.MarkFlagRequired("dataType")
	cmd.MarkFlagRequired("maximumAge")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newRetentionRuleCmd) newRetentionRule(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("dataType"); err == nil {
		if v != "" {
			body.Set("dataType", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "dataType", err))
	}
	if v, err := cmd.Flags().GetString("fragmentType"); err == nil {
		if v != "" {
			body.Set("fragmentType", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "fragmentType", err))
	}
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			body.Set("type", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	if v, err := cmd.Flags().GetString("source"); err == nil {
		if v != "" {
			body.Set("source", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "source", err))
	}
	if v, err := cmd.Flags().GetInt("maximumAge"); err == nil {
		body.Set("maximumAge", v)
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "maximumAge", err))
	}
	if cmd.Flags().Changed("editable") {
		if v, err := cmd.Flags().GetBool("editable"); err == nil {
			body.Set("editable", v)
		} else {
			return newUserError("Flag does not exist")
		}
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/retention/retentions", pathParameters)

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
