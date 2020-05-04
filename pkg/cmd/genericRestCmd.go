package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y-cli/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

type getGenericRestCmd struct {
	*baseCmd

	flagHost string
}

func newGetGenericRestCmd() *getGenericRestCmd {
	ccmd := &getGenericRestCmd{}

	cmd := &cobra.Command{
		Use:   "rest",
		Short: "Send generic REST request",
		Long:  `Send generic REST request`,
		Example: `
			Get a list of managed objects
			c8y rest GET /alarm/alarms

			c8y rest GET "/alarm/alarms?pageSize=10&status=ACTIVE"

			// Create a new alarm
			c8y rest POST "alarm/alarms" --data "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source={'id': '12345'}"
		`,
		RunE: ccmd.getGenericRest,
	}

	addDataFlag(cmd)
	cmd.Flags().String("file", "", "File to be uploaded as a binary")
	cmd.Flags().StringSliceP("header", "H", nil, "headers. i.e. --header \"Accept: value\"")
	cmd.Flags().String("accept", "", "accept (header)")
	cmd.Flags().String("contentType", "", "content type (header)")
	cmd.Flags().StringVar(&ccmd.flagHost, "host", "", "host to use for the rest request. If empty, then the session's host will be used")
	cmd.Flags().Bool("ignoreAcceptHeader", false, "Without the accept header")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getGenericRestCmd) getGenericRest(cmd *cobra.Command, args []string) error {
	method := "get"

	header := http.Header{}

	ignoreAcceptHeader := false
	if cmd.Flags().Changed("ignoreAcceptHeader") {
		if v, err := cmd.Flags().GetBool("ignoreAcceptHeader"); err == nil {
			ignoreAcceptHeader = v
		}
	}

	if v, err := cmd.Flags().GetString("accept"); err == nil && v != "" {
		if !ignoreAcceptHeader {
			header.Set("Accept", v)
		}
	}

	if v, err := cmd.Flags().GetString("contentType"); err == nil && v != "" {
		header.Set("Content-Type", v)
	}

	if values, err := cmd.Flags().GetStringSlice("header"); err == nil && len(values) > 0 {
		for _, v := range values {
			parts := strings.SplitN(v, ":", 2)
			if len(parts) != 2 {
				Logger.Warningf("Invalid header. %s", v)
				continue
			}
			Logger.Debugf("Setting header: name=%s, value=%s", parts[0], parts[1])
			header.Add(parts[0], strings.TrimSpace(parts[1]))
		}
	}

	// filter and selectors
	filters := getFilterFlag(cmd, "filter")

	var uri string
	if len(args) == 1 {
		uri = args[0]
	} else if len(args) > 1 {
		method = args[0]
		uri = args[1]
	}

	method = strings.ToUpper(method)

	if !(method == "GET" || method == "POST" || method == "PUT" || method == "DELETE") {
		return newUserError("Invalid method. Only GET, PUT, POST and DELETE are accepted")
	}

	if method == "GET" || method == "DELETE" {
		return n.doDataGenericRest(method, uri, header, nil, nil, ignoreAcceptHeader, globalFlagDryRun, filters)
	}

	if method == "PUT" || method == "POST" {
		data := getDataFlag(cmd)

		if !cmd.Flags().Changed(FlagDataName) && !cmd.Flags().Changed("file") {
			return newUserError("Missing required arguments. Either --data or --file are required")
		}

		// get file info
		var formData map[string]io.Reader
		if cmd.Flags().Changed("file") {
			formData = make(map[string]io.Reader)
			getFileFlag(cmd, "file", formData)
		}

		// Hide usage for system errors
		cmd.SilenceUsage = true
		return n.doDataGenericRest(method, uri, header, data, formData, ignoreAcceptHeader, globalFlagDryRun, filters)
	}
	return nil
}

func (n *getGenericRestCmd) doDataGenericRest(method string, path string, header http.Header, data map[string]interface{}, formData map[string]io.Reader, ignoreAcceptHeader, dryRun bool, filters *JSONFilters) error {
	baseURL, _ := url.Parse(path)

	var host string
	if n.flagHost != "" {
		host = n.flagHost
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
	defer cancel()
	start := time.Now()
	resp, err := client.SendRequest(
		ctx,
		c8y.RequestOptions{
			Method:       method,
			Host:         host,
			Path:         baseURL.Path,
			Query:        baseURL.RawQuery,
			Body:         data,
			FormData:     formData,
			Header:       header,
			DryRun:       dryRun,
			IgnoreAccept: ignoreAcceptHeader,
			ResponseData: nil,
		})

	Logger.Infof("Response time: %dms", int64(time.Since(start)/time.Millisecond))
	outputfile := globalFlagOutputFile

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

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON([]byte(*resp.JSONData))

		Logger.Debugf("Response is json: %v", isJSONResponse)

		if filters != nil && !globalFlagRaw && isJSONResponse {
			dataKey := ""
			if v := resp.JSON.Get("id"); !v.Exists() {
				// Find the property which is an array
				resp.JSON.ForEach(func(key, value gjson.Result) bool {
					if value.IsArray() {
						dataKey = key.String()
						return false
					}
					return true
				})
			}

			if dataKey != "" {
				Logger.Debugf("Data property: %s", dataKey)
			}
			responseText = filters.Apply(*resp.JSONData, dataKey)
		} else {
			responseText = []byte(*resp.JSONData)
		}

		Logger.Debugf("Pretty print: %v", globalFlagPrettyPrint)

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
