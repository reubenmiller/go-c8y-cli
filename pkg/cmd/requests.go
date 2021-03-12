package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"moul.io/http2curl"
)

// CommonCommandOptions control the handling of the response which are available for all commands
// which interact with the server
type CommonCommandOptions struct {
	ConfirmText    string
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
func (options CommonCommandOptions) AddQueryParameters(query *flags.QueryTemplate) {
	if query == nil {
		return
	}

	if options.CurrentPage > 0 {
		query.SetVariable("currentPage", options.CurrentPage)
	}

	if options.PageSize > 0 {
		query.SetVariable("pageSize", options.PageSize)
	}

	if options.WithTotalPages {
		query.SetVariable("withTotalPages", "true")
	}
}

func getCommonOptions(cmd *cobra.Command) (CommonCommandOptions, error) {
	options := CommonCommandOptions{}
	if v, err := getOutputFileFlag(cmd, "outputFile"); err == nil {
		options.OutputFile = v
	} else {
		return options, err
	}

	// default return property from the raw response
	options.ResultProperty = flags.GetCollectionPropertyFromAnnotation(cmd)

	// Filters and selectors
	options.Filters = getFilterFlag(cmd, "filter")

	if globalFlagPageSize > 0 && globalFlagPageSize != CumulocityDefaultPageSize {
		options.PageSize = globalFlagPageSize
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

	if globalFlagConfirmText != "" {
		options.ConfirmText = globalFlagConfirmText
	} else {
		options.ConfirmText = cmd.Short
	}

	return options, nil
}

func getIncludeAllFlag(cmd *cobra.Command, flagName string) (includeAll bool) {
	if v, flagErr := cmd.Flags().GetBool(flagName); flagErr == nil {
		includeAll = v
	}
	return
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout*1000)*time.Millisecond)
}

func processRequestAndResponse(requests []c8y.RequestOptions, commonOptions CommonCommandOptions) error {

	if len(requests) > 1 {
		return cmderrors.NewSystemError("Multiple request handling is currently not supported")
	}

	if len(requests) == 0 {
		return cmderrors.NewSystemError("At least one request should be given")
	}

	req := requests[0]

	// Modify request if special mode is being used
	if commonOptions.IncludeAll || commonOptions.TotalPages > 0 {
		if strings.Contains(req.Path, "inventory/managedObjects") {
			tempURL, _ := url.Parse("https://dummy.com?" + req.Query.(string))
			tempURL = optimizeManagedObjectsURL(tempURL, "0")
			req.Query = tempURL.RawQuery
			Logger.Infof("Optimizing inventory query. %v", req.Query)
		}
	}

	// enable return of would-be request
	req.DryRunResponse = true

	if !(req.Method == http.MethodPost || req.Method == http.MethodPut) {
		req.Body = nil
	}

	ctx, cancel := getTimeoutContext()
	defer cancel()
	start := time.Now()
	resp, err := client.SendRequest(
		ctx,
		req,
	)
	isDryRun := resp != nil && resp.Response.StatusCode == 0 && resp.Response.Request != nil

	if !isDryRun && resp != nil {
		durationMS := int64(time.Since(start) / time.Millisecond)
		Logger.Infof("Response time: %dms", durationMS)

		if activityLogger != nil && resp != nil {
			activityLogger.LogRequest(resp.Response, resp.JSON, durationMS)
		}
	}

	if ctx.Err() != nil {
		Logger.Errorf("request timed out after %.3fs", globalFlagTimeout)
	}

	if commonOptions.IncludeAll || commonOptions.TotalPages > 0 {
		if strings.Contains(req.Path, "inventory/managedObjects") {
			// TODO: Optimize implementation for inventory managed object queries to use the following
			Logger.Info("Using inventory optimized query")
			if err := fetchAllInventoryQueryResults(req, resp, commonOptions); err != nil {
				return err
			}
		} else {
			if err := fetchAllResults(req, resp, commonOptions); err != nil {
				return err
			}
		}
		return nil
	}

	_, err = processResponse(resp, err, commonOptions)
	return err
}

type RequestDetails struct {
	URL         string            `json:"url,omitempty"`
	Host        string            `json:"host,omitempty"`
	PathEncoded string            `json:"pathEncoded,omitempty"`
	Path        string            `json:"path,omitempty"`
	Query       string            `json:"query,omitempty"`
	Method      string            `json:"method,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
	Shell       string            `json:"shell,omitempty"`
	PowerShell  string            `json:"powershell,omitempty"`
}

func dumpRequest(w io.Writer, req *http.Request) {
	if out, err := httputil.DumpRequest(req, true); err == nil {
		fmt.Fprint(w, hideSensitiveInformationIfActive(fmt.Sprintf("%s", out)))
	}
}

// PrintRequestDetails prints the request to the console making it easier to extra informatino from it
func PrintRequestDetails(w io.Writer, requestOptions *c8y.RequestOptions, req *http.Request) {
	if globalFlagDryRunFormat == "dump" {
		dumpRequest(w, req)
		return
	}

	sectionLabel := color.New(color.Bold, color.FgHiCyan)
	label := color.New(color.FgHiCyan)
	value := color.New(color.FgGreen)

	if globalFlagNoColor {
		sectionLabel.DisableColor()
		label.DisableColor()
		value.DisableColor()
	}

	fullURL := req.URL.String()

	// strip headers which are not useful to anyone
	req.Header.Del("User-Agent")
	req.Header.Del("X-Application")
	headers := map[string]string{}
	for key := range req.Header {
		headers[key] = hideSensitiveInformationIfActive(req.Header.Get(key))
	}

	// body
	body := []byte{}
	bodyMap := make(map[string]interface{})
	var err error
	if req.Body != nil && (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch) {
		var buf bytes.Buffer
		bodyCopy := io.TeeReader(req.Body, &buf)
		req.Body = ioutil.NopCloser(&buf)

		peekBody := io.LimitReader(bodyCopy, 1024*1024)
		body, err = ioutil.ReadAll(peekBody)

		if err != nil {
			Logger.Warnf("Could not read body. %s", err)
			return
		}

		// try converting it to json
		err = jsonUtilities.ParseJSON(string(body), bodyMap)

		if err != nil {
			Logger.Warnf("Could not parse json body. %s", err)
		}
	}

	shell, pwsh, _ := getCurlCommands(req)

	details := &RequestDetails{
		URL:         fullURL,
		Host:        req.URL.Scheme + "://" + req.URL.Hostname(),
		PathEncoded: strings.Replace(fullURL, req.URL.Scheme+"://"+req.URL.Hostname(), "", 1),
		Method:      req.Method,
		Headers:     headers,
		Query:       tryUnescapeURL(req.URL.RawQuery),
		Body:        bodyMap,
		Shell:       shell,
		PowerShell:  pwsh,
	}
	details.Path = req.URL.Path

	if globalFlagDryRunFormat == "json" {
		out, err := json.Marshal(details)
		if err != nil {
			return
		}
		if !globalFlagCompact {
			out = pretty.Pretty(out)
		}
		if !globalFlagNoColor {
			out = pretty.Color(out, pretty.TerminalStyle)
		}
		// note: include newline to make it easier to parse multiple dry outputs
		fmt.Fprintf(w, "%s\n", out)
		return
	}

	if globalFlagDryRunFormat == "curl" {
		sectionLabel.Fprintf(w, "##### Curl (shell)\n\n")
		label.Fprintf(w, "```sh\n%s\n```\n", details.Shell)

		sectionLabel.Fprintf(w, "\n##### Curl (PowerShell)\n\n")
		label.Fprintf(w, "```powershell\n%s\n```\n", details.PowerShell)
		return
	}

	// markdown
	sectionLabel.Fprintf(w, "What If: Sending [%s] request to [%s]\n", req.Method, req.URL)
	label.Fprintf(w, "\n### %s %s", details.Method, tryUnescapeURL(details.PathEncoded))

	if len(req.Header) > 0 {
		// sort header names
		headerNames := make([]string, 0, len(req.Header))

		maxWidth := 0
		for key := range req.Header {
			headerNames = append(headerNames, key)
			if len(key) > maxWidth {
				maxWidth = len(key)
			}
		}

		sort.Strings(headerNames)

		label.Fprintf(w, "\n\n| %-18s| %s\n", "header", "value")
		label.Fprintf(w, "|%s|---------------------------\n", strings.Repeat("-", 19))

		for _, key := range headerNames {
			val := req.Header[key]
			label.Fprintf(w, "| %-17s | %s \n", key, hideSensitiveInformationIfActive(val[0]))
		}
	}

	if len(body) > 0 {
		if err != nil {
			Logger.Warning("failed to read all body contents")
		} else {
			sectionLabel.Fprint(w, "\n#### Body\n")
			fmt.Fprintf(w, "\n```json\n")

			if !globalFlagCompact {
				body = pretty.Pretty(body)
			}
			if !globalFlagNoColor {
				body = pretty.Color(body, pretty.TerminalStyle)
			}
			fmt.Fprintf(w, "%s", body)
			fmt.Fprintf(w, "```\n")
		}

	}
}

func tryUnescapeURL(v string) string {
	unescapedQuery, err := url.QueryUnescape(v)
	if err != nil {
		return v
	}
	return unescapedQuery
}

func getCurlCommands(req *http.Request) (shell string, pwsh string, err error) {
	if !strings.Contains("POST PUT", req.Method) {
		req.Body = nil
	}
	var command *http2curl.CurlCommand
	command, err = http2curl.GetCurlCommand(req)

	if err != nil {
		Logger.Warningf("failed to get curl command. %s", err)
		return
	}
	curlCmd := command.String()
	curlCmd = strings.ReplaceAll(curlCmd, "\n", "")

	shell = hideSensitiveInformationIfActive(curlCmd)
	pwsh = hideSensitiveInformationIfActive(strings.ReplaceAll(curlCmd, "\"", "\\\""))
	return
}

func fetchAllResults(req c8y.RequestOptions, resp *c8y.Response, commonOptions CommonCommandOptions) error {
	if req.DryRun || (resp != nil && resp.StatusCode == 0) {
		return nil
	}

	// check if response does really contain a response
	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := processResponse(resp, nil, commonOptions)

	if processErr != nil {
		return cmderrors.NewSystemError("Failed to parse response", processErr)
	}

	results := make([]*c8y.Response, 1)
	results[0] = resp

	var err error

	// start from 1, as the first request has already been sent
	currentPage := int64(1)

	// Set default total pages (when not set)
	totalPages := int64(1000)

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

		Logger.Infof("Fetching next page (%d): %s?%s", currentPage, baseURL.Path, baseURL.RawQuery)

		curReq := c8y.RequestOptions{
			Method: "GET",
			Path:   baseURL.Path,
			Query:  baseURL.RawQuery,
			Header: req.Header.Clone(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout*1000)*time.Millisecond)
		defer cancel()
		start := time.Now()
		resp, err = client.SendRequest(
			ctx,
			curReq,
		)

		// save result
		if resp != nil {
			durationMS := int64(time.Since(start) / time.Millisecond)
			Logger.Infof("Response time: %dms", durationMS)
			activityLogger.LogRequest(resp.Response, resp.JSON, durationMS)
			totalItems, processErr = processResponse(resp, err, commonOptions)

			if processErr != nil {
				return cmderrors.NewSystemError("Failed to parse response")
			}
		} else {
			break
		}

		// Check if total results is less than the pagesize, as this saves one request
		if totalItems < commonOptions.PageSize {
			Logger.Info("Found last page")
			break
		}

		if currentPage >= totalPages {
			Logger.Infof("Max pagination reached. max pages=%d", totalPages)
			break
		}

		if globalFlagIncludeAllDelayMS > 0 {
			Logger.Infof("Pausing %d ms before next request.", globalFlagIncludeAllDelayMS)
			time.Sleep(time.Duration(globalFlagIncludeAllDelayMS) * time.Millisecond)
		}
	}

	return err
}

func fetchAllInventoryQueryResults(req c8y.RequestOptions, resp *c8y.Response, commonOptions CommonCommandOptions) error {
	if req.DryRun || (resp != nil && resp.StatusCode == 0) {
		return nil
	}

	// check if response does really contain a response
	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := processResponse(resp, nil, commonOptions)

	if processErr != nil {
		return cmderrors.NewSystemError("Failed to parse response", processErr)
	}

	results := make([]*c8y.Response, 1)
	results[0] = resp

	var err error

	// start from 1, as the first request has already been sent
	currentPage := int64(1)

	// Set default total pages (when not set)
	totalPages := int64(1000)

	if commonOptions.TotalPages > 0 {
		totalPages = commonOptions.TotalPages
	}

	originalURI := ""
	lastID := "0"

	if v := resp.JSON.Get("self"); v.Exists() && v.String() != "" {
		originalURI = v.String()
	}

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

		// get last id of the result set
		if v := resp.JSON.Get(dataProperty); v.Exists() && v.IsArray() {
			items := v.Array()
			if len(items) > 0 {
				lastID = items[len(items)-1].Get("id").String()
			}
		} else {
			break
		}

		currentPage++

		baseURL, _ := url.Parse(originalURI)
		baseURL = optimizeManagedObjectsURL(baseURL, lastID)

		Logger.Infof("Fetching next page (%d): %s?%s", currentPage, baseURL.Path, baseURL.RawQuery)

		curReq := c8y.RequestOptions{
			Method: "GET",
			Path:   baseURL.Path,
			Query:  baseURL.RawQuery,
			Header: req.Header.Clone(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(globalFlagTimeout*1000)*time.Millisecond)
		defer cancel()
		start := time.Now()
		resp, err = client.SendRequest(
			ctx,
			curReq,
		)

		// save result
		if resp != nil {
			durationMS := int64(time.Since(start) / time.Millisecond)
			Logger.Infof("Response time: %dms", durationMS)
			activityLogger.LogRequest(resp.Response, resp.JSON, durationMS)

			totalItems, processErr = processResponse(resp, err, commonOptions)

			if processErr != nil {
				return cmderrors.NewSystemError("Failed to parse response")
			}
		} else {
			break
		}

		// Check if total results is less than the pagesize, as this saves one request
		if totalItems < commonOptions.PageSize {
			Logger.Info("Found last page")
			break
		}

		if currentPage >= totalPages {
			Logger.Infof("Max pagination reached. max pages=%d", totalPages)
			break
		}

		if globalFlagIncludeAllDelayMS > 0 {
			Logger.Infof("Pausing %d ms before next request.", globalFlagIncludeAllDelayMS)
			time.Sleep(time.Duration(globalFlagIncludeAllDelayMS) * time.Millisecond)
		}
	}

	return err
}

func optimizeManagedObjectsURL(u *url.URL, lastID string) *url.URL {
	q := u.Query()
	queryName := ""
	var moQuery string
	queryNames := []string{"q", "query"}

	for _, name := range queryNames {
		if v := q.Get(name); v != "" {
			queryName = name
			moQuery = v
		}
	}

	if queryName == "" {
		return u
	}

	if lastID == "" {
		lastID = "0"
	}

	if moQuery != "" {
		queryPattern := regexp.MustCompile(`^\$filter=(.+?)\s*(\$orderby=(.+?))?$`)
		matches := queryPattern.FindStringSubmatch(moQuery)

		if len(matches) >= 3 {
			matches[1] = strings.TrimSpace(matches[1])
			if strings.HasPrefix(matches[1], "(") && strings.HasSuffix(matches[1], ")") {
				moQuery = fmt.Sprintf("$filter=(_id gt '%s' and %s) $orderby=_id asc", lastID, matches[1])
			} else {
				qpart := fmt.Sprintf("_id gt '%s'", lastID)
				if len(matches[1]) > 0 {
					qpart += " and (" + matches[1] + ")"
				}
				moQuery = fmt.Sprintf("$filter=(%s) $orderby=_id asc", qpart)
			}
		}
		q.Set(queryName, moQuery)
	}
	u.RawQuery = q.Encode()
	return u
}

func processResponse(resp *c8y.Response, respError error, commonOptions CommonCommandOptions) (int, error) {
	if resp != nil && resp.StatusCode != 0 {
		Logger.Infof("Response Content-Type: %s", resp.Header.Get("Content-Type"))
		Logger.Debugf("Response Headers: %v", resp.Header)
	}

	// Display log output in special scenarios (i.e. Delete and no Accept header), so the user gets some feedback that it did something
	if resp != nil && (resp.Request.Method == http.MethodDelete && resp.StatusCode == 204 || resp.Request.Header.Get("Accept") == "" && resp.Request.Method != http.MethodDelete && resp.StatusCode == 201) {
		Logger.Warnf("%s. method: %s, status=%s, path=%s", color.GreenString("Request successful"), resp.Request.Method, resp.Status, resp.Request.URL.Path)
	}

	// write response to file instead of to stdout
	if resp != nil && respError == nil && commonOptions.OutputFile != "" {
		newline := strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json")
		fullFilePath, err := saveResponseToFile(resp, commonOptions.OutputFile, true, newline)

		if err != nil {
			return 0, cmderrors.NewSystemError("write to file failed", err)
		}

		Logger.Infof("Saved response: %s", fullFilePath)
		fmt.Printf("%s\n", fullFilePath)
		return 0, nil
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

			// Detect view (if no filters are given)
			if len(commonOptions.Filters.Pluck) == 0 {
				if resp.JSON != nil && rootCmd.dataView != nil {
					inputData := resp.JSON
					if dataProperty != "" {
						subpro := resp.JSON.Get(dataProperty)
						inputData = &subpro
					}
					props, err := rootCmd.dataView.GetView(inputData, resp.Header.Get("Content-Type"))

					if err != nil {
						Logger.Warnf("Failed to detect view. %s", err)
					} else {
						Logger.Infof("Detected data view. %s", props)
						commonOptions.Filters.Pluck = props
					}
				}
			}

			responseText = commonOptions.Filters.Apply(*resp.JSONData, dataProperty, false)

			emptyArray := []byte("[]\n")

			if len(responseText) == len(emptyArray) && bytes.Equal(responseText, emptyArray) {
				Logger.Info("No matching results found. Empty response will be ommitted")
				responseText = []byte{}
			}
		} else {
			responseText = []byte(*resp.JSONData)
		}

		if respError == nil {
			jsonformatter.WithOutputFormatters(
				Console,
				responseText,
				!isJSONResponse,
				jsonformatter.WithTrimSpace(true),
				jsonformatter.WithJSONStreamOutput(isJSONResponse, globalFlagStream, Console.IsCSV()),
				jsonformatter.WithSuffix(len(responseText) > 0, "\n"),
			)
		}
	}

	color.Unset()

	if respError != nil {
		return unfilteredSize, cmderrors.NewServerError(resp, respError)
	}
	return unfilteredSize, nil
}

func guessDataProperty(resp *c8y.Response) string {
	property := ""
	totalKeys := 0

	if v := resp.JSON.Get("id"); !v.Exists() {
		// Find the property which is an array
		resp.JSON.ForEach(func(key, value gjson.Result) bool {
			totalKeys++
			if value.IsArray() {
				property = key.String()
				return false
			}
			return true
		})
	}

	// if total keys is a high number, than it is most likely not an array of data
	// i.e. for the /tenant/statistics
	if property != "" && totalKeys > 10 {
		return ""
	}

	if property != "" && totalKeys < 10 {
		Logger.Debugf("Data property: %s", property)
	}
	return property
}

// WriteJSONToConsole writes given json output to the console supporting the common options of select, output etc.
func WriteJSONToConsole(cmd *cobra.Command, property string, output []byte) error {
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}
	output = commonOptions.Filters.Apply(string(output), property, false)

	jsonformatter.WithOutputFormatters(
		Console,
		output,
		false,
		jsonformatter.WithTrimSpace(true),
		jsonformatter.WithJSONStreamOutput(true, globalFlagStream, Console.IsCSV()),
		jsonformatter.WithSuffix(len(output) > 0, "\n"),
	)
	return nil
}
