package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

// CommonCommandOptions control the handling of the response which are available for all commands
// which interact with the server
type CommonCommandOptions struct {
	OutputFile     string
	Filters        *JSONFilters
	ResultProperty string
	IncludeAll     bool
	WithTotalPages bool
	PageSize       int
	CurrentPage    int64
	TotalPages     int64
}

// AddQueryParameters adds the common query parameters to the given query values
func (options CommonCommandOptions) AddQueryParameters(query *url.Values) {
	if query == nil {
		return
	}

	if options.CurrentPage > 0 {
		query.Add("currentPage", fmt.Sprintf("%d", options.CurrentPage))
	}

	if options.PageSize > 0 {
		if query.Get("pageSize") != "" {
			query.Set("pageSize", fmt.Sprintf("%d", options.PageSize))
		} else {
			query.Add("pageSize", fmt.Sprintf("%d", options.PageSize))
		}
	}

	if options.WithTotalPages {
		query.Add("withTotalPages", "true")
	}
}

func getCommonOptions(cmd *cobra.Command) (CommonCommandOptions, error) {
	options := CommonCommandOptions{}
	if v, err := getOutputFileFlag(cmd, "outputFile"); err == nil {
		options.OutputFile = v
	} else {
		return options, err
	}

	// Filters and selectors
	options.Filters = getFilterFlag(cmd, "filter")

	if cmd.Flags().Changed("pageSize") || globalUseNonDefaultPageSize {
		if globalFlagPageSize > 0 {
			options.PageSize = globalFlagPageSize
		}
	}

	if cmd.Flags().Changed("withTotalPages") {
		if v, err := cmd.Flags().GetBool("withTotalPages"); err == nil && v {
			options.WithTotalPages = true
		}
	}

	options.IncludeAll = getIncludeAllFlag(cmd, "includeAll")

	if options.IncludeAll {
		options.PageSize = globalFlagIncludeAllPageSize
		Logger.Debugf("Setting pageSize to maximum value to limit number of requests. value=%d", options.PageSize)
	}

	options.CurrentPage = globalFlagCurrentPage
	options.TotalPages = globalFlagTotalPages
	// options.CurrentPage = getCurrentPageFlag(cmd, "currentPage")
	// options.MaximumPage = getCurrentPageFlag(cmd, "maximumPage")

	return options, nil
}

func getIncludeAllFlag(cmd *cobra.Command, flagName string) (includeAll bool) {
	if v, flagErr := cmd.Flags().GetBool(flagName); flagErr == nil {
		includeAll = v
	}
	return
}

func getCurrentPageFlag(cmd *cobra.Command, flagName string) (currentPage int64) {
	if v, flagErr := cmd.Flags().GetInt64(flagName); flagErr == nil {
		currentPage = v
	}
	return
}

func processRequestAndResponse(requests []c8y.RequestOptions, commonOptions CommonCommandOptions) error {

	if len(requests) > 1 {
		return newSystemError("Multiple request handling is currently not supported")
	}

	if len(requests) == 0 {
		return newSystemError("At least one request should be given")
	}

	req := requests[0]

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

	if commonOptions.IncludeAll || commonOptions.TotalPages > 0 {
		if err := fetchAllResults(req, resp, commonOptions); err != nil {
			return err
		}
		return nil
	}

	_, err = processResponse(resp, err, commonOptions)
	return err
}

func fetchAllResults(req c8y.RequestOptions, resp *c8y.Response, commonOptions CommonCommandOptions) error {

	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := processResponse(resp, nil, commonOptions)

	if processErr != nil {
		return newSystemError("Failed to parse response", processErr)
	}

	results := make([]*c8y.Response, 1)
	results[0] = resp

	if resp.JSONData != nil {
		// fmt.Printf("%s\n", *resp.JSONData)
	}

	var err error

	// start from 1, as the first request has already been sent
	currentPage := int64(1)

	// Set default total pages (when not set)
	totalPages := int64(100)

	if commonOptions.TotalPages > 0 {
		totalPages = commonOptions.TotalPages
	}

	var nextURI string

	// base selection on first response
	dataProperty := guessDataProperty(resp)
	if dataProperty != "" {
		commonOptions.ResultProperty = dataProperty
	}

	// check if data is already fetched

	for {

		if resp == nil || totalItems == 0 {
			break
		}
		if v := resp.JSON.Get("next"); v.Exists() && v.String() != "" {
			nextURI = v.String()
		} else {
			break
		}

		currentPage++

		baseURL, _ := url.Parse(nextURI)

		Logger.Infof("Fetching next page: %s?%s", baseURL.Path, baseURL.RawQuery)

		curReq := c8y.RequestOptions{
			Method: "GET",
			Path:   baseURL.Path,
			Query:  baseURL.RawQuery,
			Header: req.Header.Clone(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout)*time.Millisecond)
		defer cancel()
		// start := time.Now()
		resp, err = client.SendRequest(
			ctx,
			curReq,
		)

		// save result
		if resp != nil {

			totalItems, processErr = processResponse(resp, err, commonOptions)

			if processErr != nil {
				return newSystemError("Failed to parse response")
			}
		} else {
			break
		}

		if currentPage >= totalPages {
			Logger.Infof("Max pagination reached. max pages=%d", totalPages)
			break
		}

		// time.Sleep(1000 * time.Millisecond)
	}

	return err
}

func processResponse(resp *c8y.Response, respError error, commonOptions CommonCommandOptions) (int, error) {
	// Check if pagination shou
	if resp != nil {
		Logger.Infof("Response header: %v", resp.Header)
	}

	// write response to file instead of to stdout
	if resp != nil && respError == nil && commonOptions.OutputFile != "" {
		fullFilePath, err := saveResponseToFile(resp, commonOptions.OutputFile, true)

		if err != nil {
			return 0, newSystemError("write to file failed", err)
		}

		fmt.Printf("%s\n", fullFilePath)
		return 0, nil
	}

	if resp != nil && respError == nil && resp.Header.Get("Content-Type") == "application/octet-stream" && resp.JSONData != nil {
		if encoding.IsUTF16(*resp.JSONData) {
			if utf8, err := encoding.DecodeUTF16([]byte(*resp.JSONData)); err == nil {
				fmt.Printf("%s", utf8)
			} else {
				fmt.Printf("%s\n", *resp.JSONData)
			}
		} else {
			fmt.Printf("%s\n", *resp.JSONData)
		}
		return 0, nil
	}

	if respError != nil {
		color.Set(color.FgRed, color.Bold)
	}

	unfilteredSize := 0

	if resp != nil && resp.JSONData != nil {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		Logger.Printf("Response Length: %0.1fKB", float64(len(*resp.JSONData)*1)/1024)

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON([]byte(*resp.JSONData))

		dataProperty := ""
		showRaw := globalFlagRaw || globalFlagWithTotalPages

		dataProperty = commonOptions.ResultProperty
		if dataProperty == "" {
			dataProperty = guessDataProperty(resp)
		}

		if v := resp.JSON.Get(dataProperty); v.Exists() && v.IsArray() {
			unfilteredSize = len(v.Array())
			Logger.Infof("Unfiltered array size. len=%d", unfilteredSize)
		}

		if isJSONResponse && commonOptions.Filters != nil {
			if showRaw {
				dataProperty = ""
			}
			responseText = commonOptions.Filters.Apply(*resp.JSONData, dataProperty)

			emptyArray := []byte("[]\n")

			if len(responseText) == len(emptyArray) && bytes.Compare(responseText, emptyArray) == 0 {
				Logger.Info("No matching results found. Empty response will be ommitted")
				responseText = []byte{}
			}
		} else {
			responseText = []byte(*resp.JSONData)
		}

		if globalFlagPrettyPrint && isJSONResponse {
			fmt.Printf("%s\n", pretty.Pretty(responseText))
		} else {
			fmt.Printf("%s\n", responseText)
		}
	}

	color.Unset()

	if respError != nil {
		return unfilteredSize, newSystemError("command failed", respError)
	}
	return unfilteredSize, nil
}

func guessDataProperty(resp *c8y.Response) string {
	property := ""
	if v := resp.JSON.Get("id"); !v.Exists() {
		// Find the property which is an array
		resp.JSON.ForEach(func(key, value gjson.Result) bool {
			if value.IsArray() {
				property = key.String()
				return false
			}
			return true
		})
	}

	if property != "" {
		Logger.Debugf("Data property: %s", property)
	}
	return property
}
