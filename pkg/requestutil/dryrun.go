// Package requestutil holds the request helpers of the CLI which do not depend
// on an SDK client: the dry run rendering of an HTTP request, and small pure
// helpers for requests and responses.
package requestutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/curly"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/tidwall/pretty"
)

// Printer renders a prepared HTTP request for a dry run, in the format of the
// --dryFormat option (markdown, json, dump or curl)
type Printer struct {
	Config *config.Config
	Logger *logger.Logger

	// HideSensitive hides the sensitive session values in a message (if the configuration says so)
	HideSensitive func(string) string
}

// ClientAcceptEncoding is the Accept-Encoding header which the HTTP client
// (resty) sets on a request that has none. It is removed from the requests
// (the transport then asks for "gzip", as v2.59.0 did) and from the dry run.
const ClientAcceptEncoding = "gzip, deflate"

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

func (r *Printer) DumpRequest(w io.Writer, req *http.Request) {
	if out, err := httputil.DumpRequest(req, true); err == nil {
		fmt.Fprint(w, r.HideSensitive(string(out)))
	}
}

// PrintRequestDetails prints the request to the console making it easier to extra information from it
func (r *Printer) PrintRequestDetails(w io.Writer, req *http.Request) {
	format := r.Config.DryRunFormat()

	// the HTTP client adds Accept-Encoding when the request is sent; it is
	// not part of the request the command built, so do not show it (v2.59.0
	// never did). A value the user gave (--header) is shown, as in v2.59.0.
	// Display only: the real request is unaffected.
	if req.Header.Get("Accept-Encoding") == ClientAcceptEncoding {
		req.Header.Del("Accept-Encoding")
	}

	if format == "dump" {
		// v2.59.0 dumped the User-Agent and X-Application headers of a request,
		// but not of a multipart upload (which it built without them)
		if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
			req.Header.Del("User-Agent")
			req.Header.Del("X-Application")
		}
		r.DumpRequest(w, req)
		return
	}

	// strip headers which are not useful to anyone
	req.Header.Del("User-Agent")
	req.Header.Del("X-Application")

	sectionLabel := color.New(color.Bold, color.FgHiCyan)
	label := color.New(color.FgHiCyan)
	value := color.New(color.FgGreen)

	if r.Config.DisableColor() {
		sectionLabel.DisableColor()
		label.DisableColor()
		value.DisableColor()
	}

	fullURL := req.URL.String()

	headers := map[string]string{}
	for key := range req.Header {
		headers[key] = r.HideSensitive(req.Header.Get(key))
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

	shell, pwsh, dummyFiles, _ := r.GetCurlCommands(req)

	details := &RequestDetails{
		URL:         fullURL,
		Host:        req.URL.Scheme + "://" + req.URL.Host, // Include host port number
		PathEncoded: strings.Replace(fullURL, req.URL.Scheme+"://"+req.URL.Host, "", 1),
		Method:      req.Method,
		Headers:     headers,
		Query:       TryUnescapeURL(req.URL.RawQuery),
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
			label.Fprintf(w, "| %-17s | %s \n", key, r.HideSensitive(val[0]))
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

func (r *Printer) GetCurlCommands(req *http.Request) (shell string, pwsh string, dummyFiles *curly.DummyFileCollection, err error) {
	if !strings.Contains("POST PUT DELETE", req.Method) {
		req.Body = nil
	}

	curlCmd, dummyFiles, err := curly.ToCurl(req)
	if err != nil {
		r.Logger.Warningf("failed to get curl command. %s", err)
		return
	}
	curlCmd = strings.ReplaceAll(curlCmd, "\n", "")

	shell = r.HideSensitive(curlCmd)
	pwsh = r.HideSensitive(strings.ReplaceAll(curlCmd, "\"", "\\\""))
	return
}
