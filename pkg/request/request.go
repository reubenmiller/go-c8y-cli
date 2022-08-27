package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"moul.io/http2curl/v2"
)

// Check if method supports a body with the request
func RequestSupportsBody(method string) bool {
	return c8y.RequestSupportsBody(method)
}

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
	return context.WithTimeout(context.Background(), r.Config.RequestTimeout())
}

func (r *RequestHandler) ProcessRequestAndResponse(requests []c8y.RequestOptions, commonOptions config.CommonCommandOptions) (*c8y.Response, error) {

	if len(requests) > 1 {
		return nil, cmderrors.NewSystemError("Multiple request handling is currently not supported")
	}

	if len(requests) == 0 {
		return nil, cmderrors.NewSystemError("At least one request should be given")
	}

	req := requests[0]

	// Modify request if special mode is being used
	if commonOptions.IncludeAll || commonOptions.TotalPages > 0 {
		if isInventoryQuery(&req) {
			tempURL, _ := url.Parse("https://dummy.com?" + req.Query.(string))
			tempURL = optimizeManagedObjectsURL(tempURL, "0")
			req.Query = tempURL.RawQuery
			r.Logger.Infof("Optimizing inventory query. %v", req.Query)
		}
	}

	// enable return of would-be request
	req.DryRunResponse = true

	if !RequestSupportsBody(req.Method) {
		req.Body = nil
	}

	ctx, cancel := r.GetTimeoutContext()
	ctx = context.WithValue(
		ctx,
		c8y.GetContextCommonOptionsKey(),
		c8y.CommonOptions{
			DryRun: req.DryRun,
			OnResponse: func(response *http.Response) io.Reader {
				// Add progress bar for binary downloads
				prog := r.IO.ProgressIndicator()
				if prog != nil && response.Header.Get("Content-Disposition") != "" {
					if !strings.Contains(response.Header.Get("Content-Type"), "json") && response.ContentLength > 0 {
						return c8ybinary.CreateProxyReader(prog)(response)
					}
				}
				return response.Body
			},
		})
	defer cancel()
	start := time.Now()

	// TODO: Check if this is required
	outData := make(map[string]interface{})
	req.ResponseData = &outData
	resp, err := r.Client.SendRequest(
		ctx,
		req,
	)
	isDryRun := resp != nil && resp.Response.StatusCode == 0 && resp.Response.Request != nil

	if !isDryRun && resp != nil {
		durationMS := int64(time.Since(start) / time.Millisecond)
		r.Logger.Infof("Response time: %dms", durationMS)

		if r.ActivityLogger != nil && resp != nil {
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON(), durationMS)
		}
	}

	if ctx.Err() != nil {
		r.Logger.Errorf("request timed out after %s", r.Config.RequestTimeout())
	}

	if commonOptions.IncludeAll || commonOptions.TotalPages > 0 {
		if isInventoryQuery(&req) {
			// TODO: Optimize implementation for inventory managed object queries to use the following
			r.Logger.Info("Using inventory optimized query")
			if err := r.fetchAllInventoryQueryResults(req, resp, commonOptions); err != nil {
				return nil, err
			}
		} else {
			if err := r.fetchAllResults(req, resp, commonOptions); err != nil {
				return nil, err
			}
		}
		return resp, nil
	}

	_, err = r.ProcessResponse(resp, err, commonOptions)
	return resp, err
}

func isInventoryQuery(r *c8y.RequestOptions) bool {
	if r == nil {
		return false
	}
	currentQuery := ""
	switch v := r.Query.(type) {
	case string:
		currentQuery = v
	}
	if !strings.Contains(r.Path, "inventory/managedObjects") {
		return false
	}
	if values, err := url.ParseQuery(currentQuery); err == nil {
		return values.Get("q") != "" || values.Get("query") != ""
	}
	return false
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

	if !(req != nil && r.Config.ShouldUseDryRun("")) {
		return
	}
	if req == nil {
		r.Logger.Warn("Response is nil")
		return
	}
	r.PrintRequestDetails(iostream.Out, options, req)
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
	isJSON := true

	var err error
	if req.Body != nil && RequestSupportsBody(req.Method) {
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
			r.Logger.Debugf("Using non-json body. %s", err)
			requestBody = string(body)
			isJSON = false
		}
	}

	shell, pwsh, _ := r.GetCurlCommands(req)

	details := &RequestDetails{
		URL:         fullURL,
		Host:        req.URL.Scheme + "://" + req.URL.Host, // Include host port number
		PathEncoded: strings.Replace(fullURL, req.URL.Scheme+"://"+req.URL.Host, "", 1),
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
	if !strings.Contains("POST PUT DELETE", req.Method) {
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
	if req.DryRun || (resp != nil && resp.StatusCode() == 0) {
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

	// Set total pages to unlimited
	totalPages := int64(0)

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
	timeout := r.Config.RequestTimeout()

	// check if data is already fetched

	for {

		if resp == nil || totalItems == 0 {
			break
		}
		if v := resp.JSON("next"); v.Exists() && v.String() != "" {
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
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON(), durationMS)
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

		if totalPages != 0 && currentPage >= totalPages {
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
	if req.DryRun || (resp != nil && resp.StatusCode() == 0) {
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

	// Set total pages to unlimited
	totalPages := int64(0)

	if commonOptions.TotalPages > 0 {
		totalPages = commonOptions.TotalPages
	}

	originalURI := ""
	lastID := "0"

	if v := resp.JSON("self"); v.Exists() && v.String() != "" {
		originalURI = v.String()
	}

	// base selection on first response
	dataProperty := r.guessDataProperty(resp)
	if dataProperty != "" {
		commonOptions.ResultProperty = dataProperty
	}

	delayMS := r.Config.GetIncludeAllDelay()
	timeout := r.Config.RequestTimeout()

	// check if data is already fetched
	for {

		if resp == nil || totalItems == 0 {
			break
		}

		// get last id of the result set
		if v := resp.JSON(dataProperty); v.Exists() && v.IsArray() {
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
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON(), durationMS)

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

		if totalPages != 0 && currentPage >= totalPages {
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
	if resp != nil && resp.StatusCode() != 0 {
		r.Logger.Infof("Response Content-Type: %s", resp.Response.Header.Get("Content-Type"))
		r.Logger.Debugf("Response Headers: %v", resp.Header())
	}

	// Display log output in special scenarios (i.e. Delete and no Accept header), so the user gets some feedback that it did something
	if resp != nil && (resp.Response.Request.Method == http.MethodDelete && resp.StatusCode() == 204 || resp.Response.Request.Header.Get("Accept") == "" && resp.Response.Request.Method != http.MethodDelete && (resp.StatusCode() == 201 || resp.StatusCode() == 200)) {
		if r.IsTerminal && !r.Config.ShowProgress() {
			cs := r.IO.ColorScheme()

			actionText := ""
			actionColor := cs.Green
			switch resp.Response.Request.Method {
			case http.MethodDelete:
				actionText = "Deleted"
				actionColor = cs.Red
			case http.MethodPut:
				actionText = "Updated"

			case http.MethodPost:
				actionText = "Created"

			case http.MethodPatch:
				actionText = "Patched"
			default:
				actionText = resp.Response.Request.Method
			}
			fmt.Fprintf(r.IO.ErrOut, "%s %s %s => %s\n", cs.SuccessIconWithColor(actionColor), actionText, resp.Response.Request.URL.Path, resp.Status())
		}
	}

	// write response to file instead of to stdout
	if resp != nil && respError == nil && commonOptions.OutputFileRaw != "" {
		if resp.StatusCode() != 0 {
			// check if it is a dummy response (i.e. no status code)
			newline := strings.Contains(strings.ToLower(resp.Response.Header.Get("Content-Type")), "json")
			fullFilePath, err := r.saveResponseToFile(resp, commonOptions.OutputFileRaw, false, newline)

			if err != nil {
				return 0, cmderrors.NewSystemError("write to file failed", err)
			}

			r.Logger.Infof("Saved response: %s", fullFilePath)
		}
	}

	if resp != nil && respError == nil && (r.Config.IsResponseOutput() || resp.Response.Header.Get("Content-Type") == "application/octet-stream") && len(resp.Body()) > 0 {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		r.Logger.Debugf("Writing https response output")

		if resp.Response.ContentLength > -1 {
			r.Logger.Infof("Response Length: %0.1fKB", float64(resp.Response.ContentLength)/1024)
		} else {
			r.Logger.Infof("Response Length: %0.1fKB", float64(len(resp.Body()))/1024)
		}

		outputEOL := ""
		if r.IsTerminal {
			outputEOL = "\n"
		}
		out := r.IO.Out
		if encoding.IsUTF16(resp.Body()) {
			if utf8, err := encoding.DecodeUTF16(resp.Body()); err == nil {
				fmt.Fprintf(out, "%s", utf8)
			} else {
				fmt.Fprintf(out, "%s", resp.Body())
			}
		} else {
			fmt.Fprintf(out, "%s", resp.Body())
		}
		if outputEOL != "" {
			fmt.Fprint(out, outputEOL)
		}
		return 0, nil
	}

	if respError != nil {
		color.Set(color.FgRed, color.Bold)
	}

	unfilteredSize := 0

	if resp != nil && len(resp.Body()) > 0 {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		if resp.Response.ContentLength > -1 {
			r.Logger.Infof("Response Length: %0.1fKB", float64(resp.Response.ContentLength)/1024)
		} else {
			r.Logger.Infof("Response Length: %0.1fKB", float64(len(resp.Body()))/1024)
		}

		var responseText []byte
		isJSONResponse := jsonUtilities.IsValidJSON(resp.Body())

		dataProperty := ""
		showRaw := r.Config.RawOutput() || r.Config.WithTotalPages()

		dataProperty = commonOptions.ResultProperty
		if dataProperty == "" {
			dataProperty = r.guessDataProperty(resp)
		} else if dataProperty == "-" {
			dataProperty = ""
		}

		if v := resp.JSON(dataProperty); v.Exists() && v.IsArray() {
			unfilteredSize = len(v.Array())
			r.Logger.Infof("Unfiltered array size. len=%d", unfilteredSize)
		}

		if isJSONResponse && commonOptions.Filters != nil {
			if showRaw {
				dataProperty = ""
			}

			if r.Config.RawOutput() {
				r.Logger.Infof("Raw mode active. In raw mode the following settings are forced, view=off, output=json")
			}
			view := r.Config.ViewOption()
			r.Logger.Infof("View mode: %s", view)

			// Detect view (if no filters are given)
			if len(commonOptions.Filters.Pluck) == 0 {
				if len(resp.Body()) > 0 && r.DataView != nil {
					inputData := resp.JSON()
					if dataProperty != "" {
						inputData = resp.JSON(dataProperty)
					}

					switch strings.ToLower(view) {
					case config.ViewsOff:
						// dont apply a view
						if !showRaw {
							commonOptions.Filters.Pluck = []string{"**"}
						}
					case config.ViewsAuto:
						props, err := r.DataView.GetView(&inputData, resp.Response.Header.Get("Content-Type"))

						if err != nil || len(props) == 0 {
							if err != nil {
								r.Logger.Infof("No matching view detected. defaulting to '**'. %s", err)
							} else {
								r.Logger.Info("No matching view detected. defaulting to '**'")
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

			if filterOutput, filterErr := commonOptions.Filters.Apply(string(resp.Body()), dataProperty, false, r.Console.SetHeaderFromInput); filterErr != nil {
				r.Logger.Warnf("filter error. %s", filterErr)
				responseText = filterOutput
			} else {
				responseText = filterOutput
			}

			emptyArray := []byte("[]\n")

			if len(responseText) == len(emptyArray) && bytes.Equal(responseText, emptyArray) {
				r.Logger.Info("No matching results found. Empty response will be omitted")
				responseText = []byte{}
			}
		} else {
			responseText = resp.Body()
		}

		// replace special escaped unicode sequences
		// todo: Use json encoding option (maybe in go-c8y?)
		// enc := json.NewEncoder(os.Stdout)
		// enc.SetEscapeHTML(false)
		responseText = bytes.ReplaceAll(responseText, []byte("\\u003c"), []byte("<"))
		responseText = bytes.ReplaceAll(responseText, []byte("\\u003e"), []byte(">"))
		responseText = bytes.ReplaceAll(responseText, []byte("\\u0026"), []byte("&"))

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

	if v := resp.JSON("id"); !v.Exists() {
		// Find the property which is an array
		resp.JSON().ForEach(func(key, value gjson.Result) bool {
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

	// Support simple variable substitution to be able to set the output file name dynamically to download a collection of files
	if strings.Contains(filename, "{") && strings.Contains(filename, "}") {
		if strings.Contains(filename, "{filename}") {
			if _, params, err := mime.ParseMediaType(resp.Response.Header.Get("Content-Disposition")); err == nil {
				if name, ok := params["filename"]; ok {
					filename = strings.ReplaceAll(filename, "{filename}", name)
				}
			}
		}

		if strings.Contains(filename, "{basename}") {
			if resp.Response.Request != nil {
				filename = strings.ReplaceAll(filename, "{basename}", path.Base(resp.Response.Request.URL.Path))
			}
		}

		if strings.Contains(filename, "{id}") {
			if resp.Response.Request != nil {
				r.Logger.Infof("Request: %s", resp.Response.Request.URL.Path)

				urlParts := strings.Split(resp.Response.Request.URL.Path, "/")
				for _, part := range urlParts {
					if part != "" && c8y.IsID(part) {
						r.Logger.Debugf("Found id like value. Substituting {id} for %s", part)
						filename = strings.ReplaceAll(filename, "{id}", part)
						break
					}
				}
			} else {
				r.Logger.Infof("Request is nill")
			}
		}
	}

	var out *os.File
	var err error
	dirPath := path.Dir(filename)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create directory. dir=%s,  err=%w", dirPath, err)
	}
	if append {
		out, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		out, err = os.Create(filename)
	}

	if err != nil {
		return "", fmt.Errorf("Could not create file. %s", err)
	}
	defer out.Close()

	if append && newline {
		if fs, err := out.Stat(); err == nil {
			if fs.Size() > 0 {
				// add newline when appending so that content is separated (only if file is not empty)
				fmt.Fprintf(out, "\n")
			}
		}
	}

	// Writer the body to file
	r.Logger.Printf("header: %v", resp.Header())
	fmt.Fprintf(out, "%s", resp.Body())

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
