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

type updateRetentionRuleCmd struct {
	*baseCmd
}

func newUpdateRetentionRuleCmd() *updateRetentionRuleCmd {
	ccmd := &updateRetentionRuleCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update retention rule",
		Long: `Update an existing retentule rule, i.e. change maximum number of days or the data type.
`,
		Example: `
$ c8y retentionRules get --id 12345
Update a retention rule
		`,
		RunE: ccmd.updateRetentionRule,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Retention rule id (required)")
	cmd.Flags().String("dataType", "", "RetentionRule will be applied to this type of documents, possible values [ALARM, AUDIT, EVENT, MEASUREMENT, OPERATION, *]. (required)")
	cmd.Flags().String("fragmentType", "", "RetentionRule will be applied to documents with fragmentType.")
	cmd.Flags().String("type", "", "RetentionRule will be applied to documents with type.")
	cmd.Flags().String("source", "", "RetentionRule will be applied to documents with source.")
	cmd.Flags().Int("maximumAge", 0, "Maximum age of document in days.")
	cmd.Flags().Bool("editable", false, "Whether the rule is editable. Can be updated only by management tenant.")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("dataType")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *updateRetentionRuleCmd) updateRetentionRule(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("id"); err == nil {
		if v != "" {
			pathParameters["id"] = v
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "id", err))
	}

	path := replacePathParameters("/retention/retentions/{id}", pathParameters)

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
