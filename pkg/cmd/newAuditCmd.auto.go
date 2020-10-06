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

type newAuditCmd struct {
	*baseCmd
}

func newNewAuditCmd() *newAuditCmd {
	ccmd := &newAuditCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new audit record",
		Long:  `Create a new audit record for a given action`,
		Example: `
$ c8y auditRecords create --type "ManagedObject" --time "0s" --text "Managed Object updated: my_Prop: value" --source $Device.id --activity "Managed Object updated" --severity "information"
Create an audit record for a custom managed object update
		`,
		RunE: ccmd.newAudit,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "Identifies the type of this audit record. (required)")
	cmd.Flags().String("time", "0s", "Time of the audit record.")
	cmd.Flags().String("text", "", "Text description of the audit record. (required)")
	cmd.Flags().String("source", "", "An optional ManagedObject that the audit record originated from (required)")
	cmd.Flags().String("activity", "", "The activity that was carried out. (required)")
	cmd.Flags().String("severity", "", "The severity of action: critical, major, minor, warning or information. (required)")
	cmd.Flags().String("user", "", "The user responsible for the audited action.")
	cmd.Flags().String("application", "", "The application used to carry out the audited action.")
	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("text")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("activity")
	cmd.MarkFlagRequired("severity")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newAuditCmd) newAudit(cmd *cobra.Command, args []string) error {

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
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			body.Set("type", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	if flagVal, err := cmd.Flags().GetString("time"); err == nil && flagVal != "" {
		if v, err := tryGetTimestampFlag(cmd, "time"); err == nil && v != "" {
			body.Set("time", decodeC8yTimestamp(v))
		} else {
			return newUserError("invalid date format", err)
		}
	}
	if v, err := cmd.Flags().GetString("text"); err == nil {
		if v != "" {
			body.Set("text", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "text", err))
	}
	if v, err := cmd.Flags().GetString("source"); err == nil {
		if v != "" {
			body.Set("source.id", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "source", err))
	}
	if v, err := cmd.Flags().GetString("activity"); err == nil {
		if v != "" {
			body.Set("activity", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "activity", err))
	}
	if v, err := cmd.Flags().GetString("severity"); err == nil {
		if v != "" {
			body.Set("severity", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "severity", err))
	}
	if v, err := cmd.Flags().GetString("user"); err == nil {
		if v != "" {
			body.Set("user", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "user", err))
	}
	if v, err := cmd.Flags().GetString("application"); err == nil {
		if v != "" {
			body.Set("application", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "application", err))
	}
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/audit/auditRecords", pathParameters)

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
