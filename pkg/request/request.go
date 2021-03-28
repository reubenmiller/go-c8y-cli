package request

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
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"moul.io/http2curl"
)

type RequestHandler struct {
	Console *console.Console

	IsTerminal     bool
	IO             *iostreams.IOStreams
	Client         *c8y.Client
	Config         *config.Config
	Logger         *logger.Logger
	DataView       *dataview.DataView
	ActivityLogger *activitylogger.ActivityLogger
	HideSensitive  func(*c8y.Client, string) string
}

func (r *RequestHandler) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(r.Config.RequestTimeout()*1000)*time.Millisecond)
}

func (r *RequestHandler) ProcessRequestAndResponse(requests []c8y.RequestOptions, commonOptions config.CommonCommandOptions) error {

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
			r.Logger.Infof("Optimizing inventory query. %v", req.Query)
		}
	}

	// enable return of would-be request
	req.DryRunResponse = true

	if !(req.Method == http.MethodPost || req.Method == http.MethodPut) {
		req.Body = nil
	}

	ctx, cancel := r.GetTimeoutContext()
	defer cancel()
	start := time.Now()
	resp, err := r.Client.SendRequest(
		ctx,
		req,
	)
	isDryRun := resp != nil && resp.Response.StatusCode == 0 && resp.Response.Request != nil

	if !isDryRun && resp != nil {
		durationMS := int64(time.Since(start) / time.Millisecond)
		r.Logger.Infof("Response time: %dms", durationMS)

		if r.ActivityLogger != nil && resp != nil {
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON, durationMS)
		}
	}

	if ctx.Err() != nil {
		r.Logger.Errorf("request timed out after %.3fs", r.Config.RequestTimeout())
	}

	if commonOptions.IncludeAll || commonOptions.TotalPages > 0 {
		if strings.Contains(req.Path, "inventory/managedObjects") {
			// TODO: Optimize implementation for inventory managed object queries to use the following
			r.Logger.Info("Using inventory optimized query")
			if err := r.fetchAllInventoryQueryResults(req, resp, commonOptions); err != nil {
				return err
			}
		} else {
			if err := r.fetchAllResults(req, resp, commonOptions); err != nil {
				return err
			}
		}
		return nil
	}

	_, err = r.ProcessResponse(resp, err, commonOptions)
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

func (r *RequestHandler) DumpRequest(w io.Writer, req *http.Request) {
	if out, err := httputil.DumpRequest(req, true); err == nil {
		fmt.Fprint(w, r.HideSensitive(r.Client, fmt.Sprintf("%s", out)))
	}
}

func (r *RequestHandler) DryRunHandler(iostream *iostreams.IOStreams, options *c8y.RequestOptions, req *http.Request) {

	if !r.Config.DryRun() {
		return
	}
	if req == nil {
		r.Logger.Warn("Response is nil")
		return
	}
	w := iostream.ErrOut
	if r.Config.WithError() {
		w = iostream.Out
	}

	r.PrintRequestDetails(w, options, req)
}

// PrintRequestDetails prints the request to the console making it easier to extra informatino from it
func (r *RequestHandler) PrintRequestDetails(w io.Writer, requestOptions *c8y.RequestOptions, req *http.Request) {
	format := r.Config.DryRunFormat()
	if format == "dump" {
		r.DumpRequest(w, req)
		return
	}

	sectionLabel := color.New(color.Bold, color.FgHiCyan)
	label := color.New(color.FgHiCyan)
	value := color.New(color.FgGreen)

	if r.Config.DisableColor() {
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
		headers[key] = r.HideSensitive(r.Client, req.Header.Get(key))
	}

	// body
	body := []byte{}
	var requestBody interface{}
	bodyMap := make(map[string]interface{})
	r.Logger.Warnf("input body: %s", req.Body)
	isJSON := true

	var err error
	if req.Body != nil && (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch) {
		var buf bytes.Buffer
		bodyCopy := io.TeeReader(req.Body, &buf)
		req.Body = ioutil.NopCloser(&buf)

		peekBody := io.LimitReader(bodyCopy, 1024*1024)
		body, err = ioutil.ReadAll(peekBody)

		if err != nil {
			r.Logger.Warnf("Could not read body. %s", err)
			return
		}

		// try converting it to json
		err = jsonUtilities.ParseJSON(string(body), bodyMap)

		if err == nil && (jsonUtilities.IsJSONObject(body) || jsonUtilities.IsJSONArray(body)) {
			requestBody = bodyMap
		} else {
			r.Logger.Warnf("Could not parse json body. %s", err)
			requestBody = string(body)
			isJSON = false
		}
	}
	if err != nil {
		r.Logger.Warningf("failed to read all body contents. %s", err)
	}

	shell, pwsh, _ := r.GetCurlCommands(req)

	details := &RequestDetails{
		URL:         fullURL,
		Host:        req.URL.Scheme + "://" + req.URL.Hostname(),
		PathEncoded: strings.Replace(fullURL, req.URL.Scheme+"://"+req.URL.Hostname(), "", 1),
		Method:      req.Method,
		Headers:     headers,
		Query:       tryUnescapeURL(req.URL.RawQuery),
		Body:        requestBody,
		Shell:       shell,
		PowerShell:  pwsh,
	}
	details.Path = req.URL.Path
	compactJSON := r.Config.CompactJSON()

	if format == "json" {
		out, err := json.Marshal(details)
		if err != nil {
			return
		}
		if !compactJSON {
			out = pretty.Pretty(out)
		}
		if !r.Config.DisableColor() {
			out = pretty.Color(out, pretty.TerminalStyle)
		}
		// note: include newline to make it easier to parse multiple dry outputs
		fmt.Fprintf(w, "%s\n", out)
		return
	}

	if format == "curl" {
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
			label.Fprintf(w, "| %-17s | %s \n", key, r.HideSensitive(r.Client, val[0]))
		}
	}

	if len(body) > 0 {
		if isJSON {
			sectionLabel.Fprint(w, "\n#### Body\n")
			fmt.Fprintf(w, "\n```json\n")

			if !compactJSON {
				body = pretty.Pretty(body)
			}
			if !r.Config.DisableColor() {
				body = pretty.Color(body, pretty.TerminalStyle)
			}
			fmt.Fprintf(w, "%s", body)
			fmt.Fprintf(w, "```\n")
		} else {
			sectionLabel.Fprint(w, "\n#### Body\n")
			fmt.Fprintf(w, "\n```text\n")
			fmt.Fprintf(w, "%s", body)
			fmt.Fprintf(w, "\n```\n")
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

func (r *RequestHandler) GetCurlCommands(req *http.Request) (shell string, pwsh string, err error) {
	if !strings.Contains("POST PUT", req.Method) {
		req.Body = nil
	}
	var command *http2curl.CurlCommand
	command, err = http2curl.GetCurlCommand(req)

	if err != nil {
		r.Logger.Warningf("failed to get curl command. %s", err)
		return
	}
	curlCmd := command.String()
	curlCmd = strings.ReplaceAll(curlCmd, "\n", "")

	shell = r.HideSensitive(r.Client, curlCmd)
	pwsh = r.HideSensitive(r.Client, strings.ReplaceAll(curlCmd, "\"", "\\\""))
	return
}

func (r *RequestHandler) fetchAllResults(req c8y.RequestOptions, resp *c8y.Response, commonOptions config.CommonCommandOptions) error {
	if req.DryRun || (resp != nil && resp.StatusCode == 0) {
		return nil
	}

	// check if response does really contain a response
	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := r.ProcessResponse(resp, nil, commonOptions)

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
	dataProperty := r.guessDataProperty(resp)
	if dataProperty != "" {
		commonOptions.ResultProperty = dataProperty
	}

	delayMS := r.Config.GetIncludeAllDelay()
	timeout := time.Duration(r.Config.RequestTimeout()*1000) * time.Millisecond

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

		r.Logger.Infof("Fetching next page (%d): %s?%s", currentPage, baseURL.Path, baseURL.RawQuery)

		curReq := c8y.RequestOptions{
			Method: "GET",
			Path:   baseURL.Path,
			Query:  baseURL.RawQuery,
			Header: req.Header.Clone(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		start := time.Now()
		resp, err = r.Client.SendRequest(
			ctx,
			curReq,
		)

		// save result
		if resp != nil {
			durationMS := int64(time.Since(start) / time.Millisecond)
			r.Logger.Infof("Response time: %dms", durationMS)
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON, durationMS)
			totalItems, processErr = r.ProcessResponse(resp, err, commonOptions)

			if processErr != nil {
				return cmderrors.NewSystemError("Failed to parse response")
			}
		} else {
			break
		}

		// Check if total results is less than the pagesize, as this saves one request
		if totalItems < commonOptions.PageSize {
			r.Logger.Info("Found last page")
			break
		}

		if currentPage >= totalPages {
			r.Logger.Infof("Max pagination reached. max pages=%d", totalPages)
			break
		}

		if delayMS > 0 {
			r.Logger.Infof("Pausing %d ms before next request.", delayMS)
			time.Sleep(time.Duration(delayMS) * time.Millisecond)
		}
	}

	return err
}

func (r *RequestHandler) fetchAllInventoryQueryResults(req c8y.RequestOptions, resp *c8y.Response, commonOptions config.CommonCommandOptions) error {
	if req.DryRun || (resp != nil && resp.StatusCode == 0) {
		return nil
	}

	// check if response does really contain a response
	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := r.ProcessResponse(resp, nil, commonOptions)

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
	dataProperty := r.guessDataProperty(resp)
	if dataProperty != "" {
		commonOptions.ResultProperty = dataProperty
	}

	delayMS := r.Config.GetIncludeAllDelay()
	timeout := time.Duration(r.Config.RequestTimeout()*1000) * time.Millisecond

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

		r.Logger.Infof("Fetching next page (%d): %s?%s", currentPage, baseURL.Path, baseURL.RawQuery)

		curReq := c8y.RequestOptions{
			Method: "GET",
			Path:   baseURL.Path,
			Query:  baseURL.RawQuery,
			Header: req.Header.Clone(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		start := time.Now()
		resp, err = r.Client.SendRequest(
			ctx,
			curReq,
		)

		// save result
		if resp != nil {
			durationMS := int64(time.Since(start) / time.Millisecond)
			r.Logger.Infof("Response time: %dms", durationMS)
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON, durationMS)

			totalItems, processErr = r.ProcessResponse(resp, err, commonOptions)

			if processErr != nil {
				return cmderrors.NewSystemError("Failed to parse response")
			}
		} else {
			break
		}

		// Check if total results is less than the pagesize, as this saves one request
		if totalItems < commonOptions.PageSize {
			r.Logger.Info("Found last page")
			break
		}

		if currentPage >= totalPages {
			r.Logger.Infof("Max pagination reached. max pages=%d", totalPages)
			break
		}

		if delayMS > 0 {
			r.Logger.Infof("Pausing %d ms before next request.", delayMS)
			time.Sleep(time.Duration(delayMS) * time.Millisecond)
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

func (r *RequestHandler) ProcessResponse(resp *c8y.Response, respError error, commonOptions config.CommonCommandOptions) (int, error) {
	if resp != nil && resp.StatusCode != 0 {
		r.Logger.Infof("Response Content-Type: %s", resp.Header.Get("Content-Type"))
		r.Logger.Debugf("Response Headers: %v", resp.Header)
	}

	// Display log output in special scenarios (i.e. Delete and no Accept header), so the user gets some feedback that it did something
	if resp != nil && (resp.Request.Method == http.MethodDelete && resp.StatusCode == 204 || resp.Request.Header.Get("Accept") == "" && resp.Request.Method != http.MethodDelete && resp.StatusCode == 201) {
		if r.IsTerminal {
			cs := r.IO.ColorScheme()
			fmt.Fprintf(r.IO.ErrOut, "%s %s %s => %s\n", cs.SuccessIconWithColor(cs.Red), "Deleted", resp.Request.URL.Path, resp.Status)
		}
	}

	// write response to file instead of to stdout
	if resp != nil && respError == nil && commonOptions.OutputFileRaw != "" {
		if resp.StatusCode != 0 {
			// check if it is a dummy reseponse (i.e. no status code)
			newline := strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json")
			fullFilePath, err := r.saveResponseToFile(resp, commonOptions.OutputFileRaw, false, newline)

			if err != nil {
				return 0, cmderrors.NewSystemError("write to file failed", err)
			}

			r.Logger.Infof("Saved response: %s", fullFilePath)
		}
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
		r.Logger.Printf("Response Length: %0.1fKB", float64(len(*resp.JSONData)*1)/1024)

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON([]byte(*resp.JSONData))

		dataProperty := ""
		showRaw := r.Config.RawOutput() || r.Config.WithTotalPages()

		dataProperty = commonOptions.ResultProperty
		if dataProperty == "" {
			dataProperty = r.guessDataProperty(resp)
		} else if dataProperty == "-" {
			dataProperty = ""
		}

		if v := resp.JSON.Get(dataProperty); v.Exists() && v.IsArray() {
			unfilteredSize = len(v.Array())
			r.Logger.Infof("Unfiltered array size. len=%d", unfilteredSize)
		}

		if isJSONResponse && commonOptions.Filters != nil {
			if showRaw {
				dataProperty = ""
			}

			view := r.Config.ViewOption()
			r.Logger.Infof("View mode: %s", view)

			// Detect view (if no filters are given)
			if len(commonOptions.Filters.Pluck) == 0 {
				if resp.JSON != nil && r.DataView != nil {
					inputData := resp.JSON
					if dataProperty != "" {
						subpro := resp.JSON.Get(dataProperty)
						inputData = &subpro
					}

					switch strings.ToLower(view) {
					case config.ViewsOff:
						// dont apply a view
						if !showRaw {
							commonOptions.Filters.Pluck = []string{"**"}
						}
					case config.ViewsAuto:
						props, err := r.DataView.GetView(inputData, resp.Header.Get("Content-Type"))

						if err != nil || len(props) == 0 {
							if err != nil {
								r.Logger.Infof("Failed to detect view. defaulting to '**'. %s", err)
							} else {
								r.Logger.Info("Failed to detect view. defaulting to '**'")
							}
							commonOptions.Filters.Pluck = []string{"**"}
						} else {
							r.Logger.Infof("Detected view: %s", strings.Join(props, ", "))
							commonOptions.Filters.Pluck = props
						}
					default:
						props, err := r.DataView.GetViewByName(view)
						if err != nil || len(props) == 0 {
							if err != nil {
								r.Logger.Warnf("no matching view found. %s, name=%s", err, view)
							} else {
								r.Logger.Warnf("no matching view found. name=%s", view)
							}
							commonOptions.Filters.Pluck = []string{"**"}
						} else {
							r.Logger.Infof("Detected view: %s", strings.Join(props, ", "))
							commonOptions.Filters.Pluck = props
						}
					}
				}
			} else {
				r.Logger.Debugf("using existing pluck values. %v", commonOptions.Filters.Pluck)
			}

			responseText = commonOptions.Filters.Apply(*resp.JSONData, dataProperty, false, r.Console.SetHeaderFromInput)

			emptyArray := []byte("[]\n")

			if len(responseText) == len(emptyArray) && bytes.Equal(responseText, emptyArray) {
				r.Logger.Info("No matching results found. Empty response will be ommitted")
				responseText = []byte{}
			}
		} else {
			responseText = []byte(*resp.JSONData)
		}

		consol := r.Console
		if respError == nil {
			jsonformatter.WithOutputFormatters(
				consol,
				responseText,
				!isJSONResponse,
				jsonformatter.WithFileOutput(commonOptions.OutputFile != "", commonOptions.OutputFile, false),
				jsonformatter.WithTrimSpace(true),
				jsonformatter.WithJSONStreamOutput(isJSONResponse, consol.IsJSONStream(), consol.IsCSV()),
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

func (r *RequestHandler) guessDataProperty(resp *c8y.Response) string {
	property := ""
	arrayPropertes := []string{}
	totalKeys := 0

	if v := resp.JSON.Get("id"); !v.Exists() {
		// Find the property which is an array
		resp.JSON.ForEach(func(key, value gjson.Result) bool {
			totalKeys++
			if value.IsArray() {
				arrayPropertes = append(arrayPropertes, key.String())
			}
			return true
		})
	}

	if len(arrayPropertes) > 1 {
		r.Logger.Debugf("Could not detect property as more than 1 array like property detected: %v", arrayPropertes)
		return ""
	}
	r.Logger.Debugf("Array properties: %v", arrayPropertes)

	if len(arrayPropertes) == 0 {
		return ""
	}

	property = arrayPropertes[0]

	// if total keys is a high number, than it is most likely not an array of data
	// i.e. for the /tenant/statistics
	if property != "" && totalKeys > 10 {
		return ""
	}

	if property != "" && totalKeys < 10 {
		r.Logger.Debugf("Data property: %s", property)
	}
	return property
}

// saveResponseToFile saves a response to file
// @filename	filename
// @directory	output directory. If empty, then a temp directory will be used
// if filename
func (r *RequestHandler) saveResponseToFile(resp *c8y.Response, filename string, append bool, newline bool) (string, error) {

	var out *os.File
	var err error
	if append {
		out, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		out, err = os.Create(filename)
	}

	if err != nil {
		return "", fmt.Errorf("Could not create file. %s", err)
	}
	defer out.Close()

	// Writer the body to file
	r.Logger.Printf("header: %v", resp.Header)
	_, err = io.Copy(out, resp.Body)

	if newline {
		// add trailing newline so that json lines are seperated by lines
		fmt.Fprintf(out, "\n")
	}
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents to file. %s", err)
	}

	if fullpath, err := filepath.Abs(filename); err == nil {
		return fullpath, nil
	}
	return filename, nil
}

// HasJSONHeader returns true if the header contains a json content type
func HasJSONHeader(h *http.Header) bool {
	if h == nil {
		return true
	}
	contentType := h.Get("Content-Type")
	if contentType == "" {
		return true
	}
	return strings.Contains(strings.ToLower(contentType), "json")
}
