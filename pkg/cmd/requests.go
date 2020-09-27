package cmd

import (
	"context"
	"fmt"
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

	return options, nil
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

	return processResponse(resp, err, commonOptions)
}

func processResponse(resp *c8y.Response, respError error, commonOptions CommonCommandOptions) error {
	if resp != nil {
		Logger.Infof("Response header: %v", resp.Header)
	}

	// write response to file instead of to stdout
	if resp != nil && respError == nil && commonOptions.OutputFile != "" {
		fullFilePath, err := saveResponseToFile(resp, commonOptions.OutputFile)

		if err != nil {
			return newSystemError("write to file failed", err)
		}

		fmt.Printf("%s", fullFilePath)
		return nil
	}

	if resp != nil && respError == nil && resp.Header.Get("Content-Type") == "application/octet-stream" && resp.JSONData != nil {
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

	if respError != nil {
		color.Set(color.FgRed, color.Bold)
	}

	if resp != nil && resp.JSONData != nil {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		Logger.Printf("Response Length: %0.1fKB", float64(len(*resp.JSONData)*1)/1024)

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON([]byte(*resp.JSONData))

		dataProperty := commonOptions.ResultProperty
		if dataProperty == "" {
			dataProperty = guessDataProperty(resp)
		}

		if isJSONResponse && commonOptions.Filters != nil && !globalFlagRaw {
			responseText = commonOptions.Filters.Apply(*resp.JSONData, dataProperty)
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

	if respError != nil {
		return newSystemError("command failed", respError)
	}
	return nil
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
