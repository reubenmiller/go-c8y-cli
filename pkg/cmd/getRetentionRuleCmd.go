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

type getRetentionRuleCmd struct {
	*baseCmd
}

func newGetRetentionRuleCmd() *getRetentionRuleCmd {
	ccmd := &getRetentionRuleCmd{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get retention rule",
		Long: `Get an existing retention by id
`,
		Example: `
$ c8y retentionRules get --id 12345
Get a retention rule
		`,
		RunE: ccmd.getRetentionRule,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Retention rule id (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getRetentionRuleCmd) getRetentionRule(cmd *cobra.Command, args []string) error {

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

	// filter and selectors
	filters := getFilterFlag(cmd, "filter")

	req := c8y.RequestOptions{
		Method:       "GET",
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

	return n.doGetRetentionRule(req, outputfile, filters)
}

func (n *getRetentionRuleCmd) doGetRetentionRule(req c8y.RequestOptions, outputfile string, filters *JSONFilters) error {
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
