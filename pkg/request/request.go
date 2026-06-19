package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
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
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/curly"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encoding"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonformatter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
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

func (r *RequestHandler) ProcessRequestAndResponse(requests []c8y.RequestOptions, input any, commonOptions config.CommonCommandOptions) (*c8y.Response, error) {

	if len(requests) > 1 {
		return nil, cmderrors.NewSystemError("Multiple request handling is currently not supported")
	}

	if len(requests) == 0 {
		return nil, cmderrors.NewSystemError("At least one request should be given")
	}

	req := requests[0]

	// Modify request if special mode is being used
	if commonOptions.IncludeAll || commonOptions.TotalPages > 1 {
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
			DryRun:    req.DryRun,
			WithError: r.Config.WithError(),
			OnResponse: func(response *http.Response) io.Reader {
				// Add progress bar for binary downloads
				prog := r.IO.ProgressIndicator()
				if prog != nil && response.Header.Get("Content-Disposition") != "" {
					if response.ContentLength > 0 {
						return c8ybinary.CreateProxyReader(prog)(response)
					}
				}
				return response.Body
			},
		})
	defer cancel()

	// Support both JSON objects or arrays, but default to an object
	if req.ResponseData == nil {
		req.ResponseData = make(map[string]interface{})
	}
	resp, err := r.Client.SendRequest(
		ctx,
		req,
	)
	isDryRun := resp != nil && resp.Response.StatusCode == 0 && resp.Response.Request != nil

	if !isDryRun && resp != nil {
		durationMS := resp.Duration().Milliseconds()
		r.Logger.Infof("Response time: %dms", durationMS)

		if r.ActivityLogger != nil && resp != nil {
			r.ActivityLogger.LogRequest(resp.Response, resp.JSON(), durationMS)
		}
	}

	if ctx.Err() != nil {
		r.Logger.Errorf("request timed out after %s", r.Config.RequestTimeout())
	}

	if commonOptions.IncludeAll || commonOptions.TotalPages > 1 {
		if isInventoryQuery(&req) {
			// TODO: Optimize implementation for inventory managed object queries to use the following
			r.Logger.Info("Using inventory optimized query")
			if err := r.fetchAllInventoryQueryResults(req, resp, input, commonOptions); err != nil {
				return nil, err
			}
		} else {
			if err := r.fetchAllResults(req, resp, input, commonOptions); err != nil {
				return nil, err
			}
		}
		return resp, nil
	}

	_, err = r.ProcessResponse(resp, err, input, commonOptions)
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
	URL         string              `json:"url,omitempty"`
	Host        string              `json:"host,omitempty"`
	PathEncoded string              `json:"pathEncoded,omitempty"`
	Path        string              `json:"path,omitempty"`
	Query       string              `json:"query,omitempty"`
	QueryParams map[string][]string `json:"queryParams,omitempty"`
	Method      string              `json:"method,omitempty"`
	Headers     map[string]string   `json:"headers,omitempty"`
	Body        interface{}         `json:"body,omitempty"`
	Shell       string              `json:"shell,omitempty"`
	PowerShell  string              `json:"powershell,omitempty"`
}

func (r *RequestHandler) DumpRequest(w io.Writer, req *http.Request) {
	if out, err := httputil.DumpRequest(req, true); err == nil {
		fmt.Fprint(w, r.HideSensitive(r.Client, string(out)))
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

// PrintRequestDetails prints the request to the console making it easier to extra information from it
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
		req.Body = io.NopCloser(&buf)

		peekBody := io.LimitReader(bodyCopy, 1024*1024)
		body, err = io.ReadAll(peekBody)

		if err != nil {
			r.Logger.Warnf("Could not read body. %s", err)
			return
		}

		// FIXME: This seems overly complicated. The json parsing should be able to handle any kind of json object
		// regardless whether it is an array, object, string, number etc.
		// try converting it to json
		if jsonUtilities.IsJSONArray(body) {
			var bodyArray []any
			if err := json.Unmarshal(body, &bodyArray); err == nil {
				requestBody = bodyArray
			} else {
				requestBody = string(body)
				isJSON = false
			}
		} else if jsonUtilities.IsJSONObject(body) {
			if err := jsonUtilities.ParseJSON(string(body), bodyMap); err == nil {
				requestBody = bodyMap
			} else {
				r.Logger.Debugf("Error parsing json object in dry run. %s", err)
				requestBody = string(body)
				isJSON = false
			}
		} else {
			r.Logger.Debugf("Using non-json body. %s", err)
			requestBody = string(body)
			isJSON = false
		}
	}

	shell, pwsh, dummyFiles, _ := r.GetCurlCommands(requestOptions, req)

	// queryParams is the decoded form of Query as a map, so consumers (e.g. test
	// assertions, --select) can address a single parameter by name instead of
	// substring-matching the raw string. It is url.Values (map[string][]string)
	// so a parameter repeated in the query — "status=ACTIVE&status=ACKNOWLEDGED"
	// — is preserved as ["ACTIVE","ACKNOWLEDGED"] rather than being collapsed,
	// which a plain map[string]string could not represent. nil (omitted) when the
	// request carries no query string.
	var queryParams map[string][]string
	if values := req.URL.Query(); len(values) > 0 {
		queryParams = values
	}

	details := &RequestDetails{
		URL:         fullURL,
		Host:        req.URL.Scheme + "://" + req.URL.Host, // Include host port number
		PathEncoded: strings.Replace(fullURL, req.URL.Scheme+"://"+req.URL.Host, "", 1),
		Method:      req.Method,
		Headers:     headers,
		Query:       TryUnescapeURL(req.URL.RawQuery),
		QueryParams: queryParams,
		Body:        requestBody,
		Shell:       shell,
		PowerShell:  pwsh,
	}
	details.Path = req.URL.Path
	compactJSON := r.Config.CompactJSON()

	if format == "json" {
		// Encode without HTML escaping so query strings keep their literal
		// '&'/'<'/'>' (e.g. "pageSize=1&withTotalPages=true") instead of
		// & etc. This matches the CLI's JSON output convention, which the
		// downstream consumers (util show, --select) expect.
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(details); err != nil {
			return
		}
		out := bytes.TrimRight(buf.Bytes(), "\n")
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
		sectionLabel.Fprintf(w, "##### curl (shell)\n\n")
		label.Fprintf(w, "```sh\n%s\n```\n", details.Shell)

		sectionLabel.Fprintf(w, "\n##### curl (PowerShell)\n\n")
		label.Fprintf(w, "```powershell\n%s\n```\n", details.PowerShell)

		if len(dummyFiles.Files) > 0 {
			for _, file := range dummyFiles.Files {
				contents := file.Contents
				if len(contents) > 1024 {
					contents = append(contents[:1024], []byte("\n...")...)
				}
				sectionLabel.Fprintf(w, "\n##### file: %s\n\n", file.Name)
				label.Fprintf(w, "\n```text/plain\n%s\n```\n", contents)
			}
		}
		return
	}

	// markdown
	sectionLabel.Fprintf(w, "What If: Sending [%s] request to [%s]\n", req.Method, req.URL)
	label.Fprintf(w, "\n### %s %s", details.Method, TryUnescapeURL(details.PathEncoded))

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

func TryUnescapeURL(v string) string {
	unescapedQuery, err := url.QueryUnescape(v)
	if err != nil {
		return v
	}
	return unescapedQuery
}

func (r *RequestHandler) GetCurlCommands(requestOptions *c8y.RequestOptions, req *http.Request) (shell string, pwsh string, dummyFiles *curly.DummyFileCollection, err error) {
	if !strings.Contains("POST PUT DELETE", req.Method) {
		req.Body = nil
	}

	curlCmd, dummyFiles, err := curly.ToCurl(req)
	if err != nil {
		r.Logger.Warningf("failed to get curl command. %s", err)
		return
	}
	curlCmd = strings.ReplaceAll(curlCmd, "\n", "")

	shell = r.HideSensitive(r.Client, curlCmd)
	pwsh = r.HideSensitive(r.Client, strings.ReplaceAll(curlCmd, "\"", "\\\""))
	return
}

func (r *RequestHandler) fetchAllResults(req c8y.RequestOptions, resp *c8y.Response, input any, commonOptions config.CommonCommandOptions) error {
	if req.DryRun || (resp != nil && resp.StatusCode() == 0) {
		return nil
	}

	// check if response does really contain a response
	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := r.ProcessResponse(resp, nil, input, commonOptions)

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
			totalItems, processErr = r.ProcessResponse(resp, err, input, commonOptions)

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

func (r *RequestHandler) fetchAllInventoryQueryResults(req c8y.RequestOptions, resp *c8y.Response, input any, commonOptions config.CommonCommandOptions) error {
	if req.DryRun || (resp != nil && resp.StatusCode() == 0) {
		return nil
	}

	// check if response does really contain a response
	if resp == nil {
		return fmt.Errorf("Response is empty")
	}

	var totalItems int

	totalItems, processErr := r.ProcessResponse(resp, nil, input, commonOptions)

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

			totalItems, processErr = r.ProcessResponse(resp, err, input, commonOptions)

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

func FlattenArrayMap[K string, V []string](m map[K]V) map[K]any {
	out := make(map[K]any)
	for key, value := range m {
		if len(value) == 1 {
			out[key] = value[0]
		} else {
			out[key] = value
		}
	}
	return out
}

func ExecuteTemplate(responseText []byte, resp *http.Response, input any, commonOptions config.CommonCommandOptions, duration time.Duration) ([]byte, error) {
	requestData := make(map[string]interface{})
	requestData["path"] = resp.Request.URL.Path
	requestData["pathEncoded"] = strings.Replace(resp.Request.URL.String(), resp.Request.URL.Scheme+"://"+resp.Request.URL.Host, "", 1)
	requestData["host"] = resp.Request.URL.Host
	requestData["url"] = resp.Request.URL.String()
	requestData["query"] = TryUnescapeURL(resp.Request.URL.RawQuery)
	requestData["queryParams"] = FlattenArrayMap(resp.Request.URL.Query())
	requestData["method"] = resp.Request.Method
	// requestData["header"] = resp.Response.Request.Header

	// TODO: Add a response variable to included the status code, content type,
	responseData := make(map[string]interface{})
	responseData["statusCode"] = resp.StatusCode
	responseData["status"] = resp.Status
	responseData["duration"] = duration.Milliseconds()
	responseData["contentLength"] = resp.ContentLength
	responseData["contentType"] = resp.Header.Get("Content-Type")
	responseData["header"] = FlattenArrayMap(resp.Header)
	responseData["proto"] = resp.Proto
	responseData["body"] = string(responseText)

	tmpl, err := mapbuilder.GetOutputTemplate(commonOptions.OutputTemplate)
	if err != nil {
		return nil, err
	}
	return tmpl.Evaluate(responseText, input, requestData, responseData, commonOptions.CommandFlags)
}

func printResponseSize(l *logger.Logger, resp *c8y.Response) {
	if resp.Response.ContentLength > -1 {
		l.Infof("Response Length: %0.1fKB", float64(resp.Response.ContentLength)/1024)
	} else {
		if resp.Response.Uncompressed {
			l.Infof("Response Length: %0.1fKB (uncompressed)", float64(len(resp.Body()))/1024)
		} else {
			l.Infof("Response Length: %0.1fKB", float64(len(resp.Body()))/1024)
		}
	}
}

func (r *RequestHandler) ProcessResponse(resp *c8y.Response, respError error, input any, commonOptions config.CommonCommandOptions) (int, error) {
	if resp != nil && resp.StatusCode() != 0 {
		r.Logger.Infof("Response Content-Type: %s", resp.Response.Header.Get("Content-Type"))
		r.Logger.Debugf("Response Headers: %v", resp.Header())
	}

	// Note: An output template will affect the handling of the response
	// for example responses with no responses will generate output due to the output template
	hasOutputTemplate := commonOptions.OutputTemplate != ""

	// Display log output in special scenarios (i.e. Delete and no Accept header), so the user gets some feedback that it did something
	if resp != nil {
		if resp.Response.Uncompressed {
			// Add Accept Encoding back in if the response was uncompressed
			// But the original Content-Length setting is still lost
			resp.Header().Set("Accept-Encoding", "gzip")
		}
		showMessage := resp.StatusCode() == 204 ||
			(resp.Response.Header.Get("Content-Type") == "" ||
				resp.Response.Request.Header.Get("Accept") == "") && resp.StatusCode() >= 200 && resp.StatusCode() < 400

		if showMessage && !hasOutputTemplate {
			if r.Config.ForceTTY() || (r.IsTerminal && !r.Config.ShowProgress()) {
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
	}

	// write response to file instead of to stdout
	if resp != nil && respError == nil && commonOptions.OutputFileRaw != "" {
		if resp.StatusCode() != 0 {
			// check if it is a dummy response (i.e. no status code)
			newline := strings.Contains(strings.ToLower(resp.Response.Header.Get("Content-Type")), "json")
			fields := make(map[string]string)
			// Only works if it is json, otherwise these values will be empty
			props := []string{"id", "name", "type", "owner"}
			for _, prop := range props {
				fields[prop] = resp.JSON(prop).String()
			}

			if _, params, err := mime.ParseMediaType(resp.Response.Header.Get("Content-Disposition")); err == nil {
				if name, ok := params["filename"]; ok {
					fields["filename"] = name
				}
			}
			if resp.Response.Request != nil {
				fields["basename"] = path.Base(resp.Response.Request.URL.Path)
			}

			if resp.Response.Request != nil {
				r.Logger.Infof("Request: %s", resp.Response.Request.URL.Path)

				for part := range strings.SplitSeq(resp.Response.Request.URL.Path, "/") {
					if part != "" && c8y.IsID(part) {
						r.Logger.Debugf("Found id like value. value=%s", part)
						fields["id"] = part
						break
					}
				}
			}

			fullFilePath, err := fileutilities.WriteToFile(bytes.NewReader(resp.Body()), commonOptions.OutputFileRaw, false, newline, fields)

			if err != nil {
				return 0, cmderrors.NewSystemError("write to file failed", err)
			}

			r.Logger.Infof("Saved response: %s", fullFilePath)
		}
	}

	if resp != nil && respError == nil && !hasOutputTemplate && (r.Config.IsResponseOutput() || resp.Response.Header.Get("Content-Type") == "application/octet-stream") && len(resp.Body()) > 0 {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		r.Logger.Debugf("Writing https response output")

		printResponseSize(r.Logger, resp)

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

	// Trim space from json object/array output
	trimSpaceFromOutput := true

	if resp != nil && (len(resp.Body()) > 0 || hasOutputTemplate) {
		// estimate size based on utf8 encoding. 1 char is 1 byte
		printResponseSize(r.Logger, resp)

		var responseText []byte

		dataProperty := commonOptions.ResultProperty

		if respError != nil && r.Config.WithError() {
			// use the raw response
			serverError := cmderrors.NewServerError(resp, respError, r.IO, !r.Config.WithError())
			resp.SetBody([]byte(serverError.JSONString()))
			dataProperty = "-"
		}

		isJSONResponse := jsonUtilities.IsValidJSON(resp.Body())

		showRaw := r.Config.RawOutput() || r.Config.WithTotalPages() || r.Config.WithTotalElements()

		if dataProperty == "" {
			dataProperty = r.guessDataProperty(resp)
		} else if dataProperty == "-" {
			dataProperty = ""
		}

		if v := resp.JSON(dataProperty); v.Exists() && v.IsArray() {
			unfilteredSize = len(v.Array())
			r.Logger.Infof("Unfiltered array size. len=%d", unfilteredSize)
		}

		// Apply output template (before the data is processed as the template can transform text to json or other way around)
		if hasOutputTemplate {
			var tempBody []byte
			if showRaw || dataProperty == "" {
				tempBody = resp.Body()
			} else {
				tempBody = []byte(resp.JSON(dataProperty).Raw)
			}
			dataProperty = ""

			tmplOutput, tmplErr := ExecuteTemplate(tempBody, resp.Response, input, commonOptions, resp.Duration())
			if tmplErr != nil {
				return unfilteredSize, tmplErr
			}

			if jsonUtilities.IsValidJSON(tmplOutput) {
				isJSONResponse = true
				resp.SetBody(pretty.Ugly(tmplOutput))
			} else {
				// Preserve output
				trimSpaceFromOutput = false
				isJSONResponse = false

				// Decode output
				var tmplOutputDecoded []byte
				if parsedOutput := gjson.ParseBytes(tmplOutput); parsedOutput.Exists() && parsedOutput.Type == gjson.String {
					tmplOutputDecoded = []byte(parsedOutput.Str)
				} else {
					tmplOutputDecoded = tmplOutput
				}

				resp.SetBody(tmplOutputDecoded)
			}
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
						// don't apply a view
						if !showRaw {
							commonOptions.Filters.Pluck = []string{"**"}
						}
					case config.ViewsAuto:
						viewData := &dataview.ViewData{
							ResponseBody: &inputData,
							ContentType:  resp.Response.Header.Get("Content-Type"),
							Request:      resp.Response.Request,
						}
						if resp.Response != nil {
							viewData.Request = resp.Response.Request
						}
						props, err := r.DataView.GetView(viewData)

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

			if !showRaw {
				if len(responseText) == len(emptyArray) && bytes.Equal(responseText, emptyArray) {
					r.Logger.Info("No matching results found. Empty response will be omitted")
					responseText = []byte{}
				}
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

		// Wait for progress bar to finish before printing to console
		// to prevent overriding the output
		r.IO.WaitForProgressIndicator()

		consol := r.Console
		if respError == nil || commonOptions.WithError {
			jsonformatter.WithOutputFormatters(
				consol,
				responseText,
				!isJSONResponse,
				jsonformatter.WithFileOutput(commonOptions.OutputFile != "", commonOptions.OutputFile, false),
				jsonformatter.WithTrimSpace(trimSpaceFromOutput),
				jsonformatter.WithJSONStreamOutput(isJSONResponse, consol.IsJSONStream(), consol.IsTextOutput()),
				jsonformatter.WithSuffix(len(responseText) > 0, "\n"),
			)
		}
	}

	color.Unset()

	if respError != nil {
		// Don't double wrap errors
		if userErr, ok := respError.(cmderrors.CommandError); ok {
			return unfilteredSize, userErr
		}
		rErr := cmderrors.NewServerError(resp, respError, r.IO, !r.Config.WithError())
		if hasOutputTemplate {
			rErr.Processed = true
		}
		return unfilteredSize, rErr
	}
	return unfilteredSize, nil
}

func (r *RequestHandler) guessDataProperty(resp *c8y.Response) string {
	property := ""
	arrayProperties := []string{}
	totalKeys := 0

	if v := resp.JSON("id"); !v.Exists() {
		// Find the property which is an array
		resp.JSON().ForEach(func(key, value gjson.Result) bool {
			totalKeys++
			if value.IsArray() {
				arrayProperties = append(arrayProperties, key.String())
			}
			return true
		})
	}

	if len(arrayProperties) > 1 {
		r.Logger.Debugf("Could not detect property as more than 1 array like property detected: %v", arrayProperties)
		return ""
	}
	r.Logger.Debugf("Array properties: %v", arrayProperties)

	if len(arrayProperties) == 0 {
		return ""
	}

	property = arrayProperties[0]

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
