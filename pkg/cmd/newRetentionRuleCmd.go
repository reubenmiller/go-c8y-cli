// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
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
	queryValue, err := url.QueryUnescape(query.Encode())

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

	// filter and selectors
	filters := getFilterFlag(cmd, "filter")

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

	// Common outputfile option
	outputfile := ""
	if v, err := getOutputFileFlag(cmd, "outputFile"); err == nil {
		outputfile = v
	} else {
		return err
	}

	return n.doNewRetentionRule(req, outputfile, filters)
}

func (n *newRetentionRuleCmd) doNewRetentionRule(req c8y.RequestOptions, outputfile string, filters *JSONFilters) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
	defer cancel()
	start := time.Now()
	resp, err := client.SendRequest(
		ctx,
		req,
	)

	Logger.Infof("Response time: %dms", int64(time.Since(start)/time.Millisecond))

	if ctx.Err() != nil {
		Logger.Criticalf("request timed out after %d", globalFlagTimeout)
	}

	if resp != nil {
		Logger.Infof("Response header: %v", resp.Header)
	}

	// write response to file instead of to stdout
	if resp != nil && err == nil && outputfile != "" {
		fullFilePath, err := saveResponseToFile(resp, outputfile)

		if err != nil {
			return newSystemError("write to file failed", err)
		}

		fmt.Printf("%s", fullFilePath)
		return nil
	}

	if resp != nil && err == nil && resp.Header.Get("Content-Type") == "application/octet-stream" && resp.JSONData != nil {
		if encoding.IsUTF16(*resp.JSONData) {
			if utf8, err := encoding.DecodeUTF16([]byte(*resp.JSONData)); err == nil {
				fmt.Printf("%s", utf8)
			} else {
				fmt.Printf("%s", *resp.JSONData)
			}
		} else {
			fmt.Printf("%s", *resp.JSONData)
		}
		return nil
	}

	if err != nil {
		color.Set(color.FgRed, color.Bold)
	}

	if resp != nil && resp.JSONData != nil {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		Logger.Printf("Response Length: %0.1fKB", float64(len(*resp.JSONData)*1)/1024)

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON([]byte(*resp.JSONData))

		if isJSONResponse && filters != nil && !globalFlagRaw {
			responseText = filters.Apply(*resp.JSONData, "")
		} else {
			responseText = []byte(*resp.JSONData)
		}

		if globalFlagPrettyPrint && isJSONResponse {
			fmt.Printf("%s", pretty.Pretty(responseText))
		} else {
			fmt.Printf("%s", responseText)
		}
	}

	color.Unset()

	if err != nil {
		return newSystemError("command failed", err)
	}
	return nil
}
