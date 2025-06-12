package registerBulk

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/randdata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type KeyValuePair struct {
	Key   string
	Value string
}

func writeCSV(w io.ReadWriter, items []KeyValuePair) {
	contents := csv.NewWriter(w)
	// Use tab delimiter to avoid csv problems when values container a comma
	contents.Comma = '\t'

	record := [2][]string{}
	for _, i := range items {
		if i.Value != "" {
			record[0] = append(record[0], i.Key)
			record[1] = append(record[1], i.Value)
		}
	}

	contents.Write(record[0])

	contents.Flush()

	s := strings.Join(record[1], string(contents.Comma))
	io.WriteString(w, s+"\n")
}

type PayloadMapping struct {
	CSVHeader  string
	Properties Value
	Output     Value
}

type Value func(gjson.Result, ...string) (string, string)

func WithValue(key ...string) Value {
	return func(result gjson.Result, other ...string) (string, string) {
		if len(other) > 0 && other[0] != "" {
			return key[0], other[0]
		}
		for _, k := range key {
			v := result.Get(k).String()
			if v != "" {
				return k, v
			}
		}
		return key[0], ""
	}
}

func WithPasswordOrDefault(key string) Value {
	return func(result gjson.Result, other ...string) (string, string) {
		credential := result.Get(key).String()
		if credential == "" {
			// Show the password to the user (as they need this when connecting the device)
			credential = randdata.Password(32)
			// showPassword = true
		}

		passwordWarningChars := "\""
		if strings.ContainsAny(credential, passwordWarningChars) {
			// llog.Warnf("Device password contains some unsupported characters [%s]. Please avoid using any of them", passwordWarningChars)
		}
		return key, credential
	}
}

func CreateCSVPayload(b io.ReadWriter, input gjson.Result, mappings []PayloadMapping) map[string]any {
	csvKeyPairs := make([]KeyValuePair, len(mappings))

	output := map[string]any{}

	for _, mapping := range mappings {
		_, value := mapping.Properties(input)
		csvKeyPairs = append(csvKeyPairs, KeyValuePair{
			Key:   mapping.CSVHeader,
			Value: value,
		})
		if mapping.Output != nil {
			outputKey, outputValue := mapping.Output(input, value)
			output[outputKey] = outputValue
		}
	}
	writeCSV(b, csvKeyPairs)

	return output
}

type RegistrationOptions struct {
	Config        *config.Config
	Log           *logger.Logger
	Client        *c8y.Client
	Factory       *cmdutil.Factory
	CommonOptions config.CommonCommandOptions
}

func RunBulkRegistrationJob(cmd *cobra.Command, opts *RegistrationOptions, mappings []PayloadMapping) func(j worker.Job) (any, error) {
	return func(j worker.Job) (any, error) {
		options := gjson.ParseBytes(j.Value.([]byte))

		formData := make(map[string]io.Reader)
		b := bytes.NewBufferString("")

		output := CreateCSVPayload(b, options, mappings)

		formData["file"] = b

		req := c8y.RequestOptions{
			Method:       http.MethodPost,
			Path:         "devicecontrol/bulkNewDeviceRequests",
			Accept:       "application/json",
			FormData:     formData,
			IgnoreAccept: opts.Config.IgnoreAcceptHeader(),
			DryRun:       opts.Config.ShouldUseDryRun(cmd.CommandPath()),
		}

		response, responseErr := opts.Client.SendRequest(context.Background(), req)
		if responseErr != nil {
			return response, responseErr
		}

		// dry run
		if response == nil || response.IsDryRun() {
			return "", nil
		}

		body := response.Body()
		totalFailed := gjson.GetBytes(body, "numberOfFailed").Int()
		if totalFailed != 0 {
			opts.Log.Infof("Response: %v", response)
			failuresReasons := make([]string, 0)
			response.JSON("failedCreationList").ForEach(func(key, value gjson.Result) bool {
				if v := value.Get("failureReason"); v.Exists() {
					if reason := v.String(); reason != "" {
						failuresReasons = append(failuresReasons, fmt.Sprintf("id=%s, reason=%s", value.Get("deviceId").String(), reason))
					}
				}
				return true
			})
			return response, cmderrors.NewUserError(fmt.Sprintf("bulk registration has some failures. failed=%d, reasons=%v", totalFailed, failuresReasons))
		}
		opts.Log.Infof("Bulk registration was successful. %v", response)

		// Lookup device id so that the command can be piped to downstream items
		identity, _, identityErr := opts.Client.Identity.GetExternalID(context.Background(), options.Get("external-type").String(), options.Get("id").String())
		if identityErr != nil {
			return "", identityErr
		}

		// Build a response to return to the user (this is not the response receive from c8y)
		output["id"] = identity.ManagedObject.ID
		if tenant := options.Get("tenant").String(); tenant != "" {
			output["username"] = fmt.Sprintf("%s/device_%s", tenant, options.Get("id").String())
		} else {
			tenant := opts.Client.GetTenantName(context.Background())
			output["username"] = fmt.Sprintf("%s/device_%s", tenant, options.Get("id").String())
		}

		outB, jsonErr := json.Marshal(output)
		if jsonErr != nil {
			return "", jsonErr
		}

		contentType := response.Response.Header.Get("Content-Type")
		opts.Log.Infof("API Content-Type: %s", contentType)

		err := opts.Factory.WriteOutput(outB, cmdutil.OutputContext{
			Input:    j.Input,
			Response: response.Response,
		}, &opts.CommonOptions)
		return nil, err
	}
}
