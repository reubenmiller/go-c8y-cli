package c8ystream

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/tidwall/gjson"
)

// RequestCall returns the Call which sends a custom request (client.SendRequest)
// and renders the response the same way for any endpoint:
//
//   - a JSON response yields the items of its collection (the response's only
//     array property, e.g. managedObjects), or the whole response when it has
//     no collection or with --raw, --withTotalPages or --withTotalElements;
//   - with --includeAll or --totalPages, a GET follows the collection's next
//     page links;
//   - any other response body (and every body with --output serverresponse)
//     is written as is to --outputFileRaw, or else to stdout;
//   - an HTTP error status is an item error.
func (r *Runner) RequestCall(client *apiv2.Client, opts apiv2.RequestOptions) Call {
	return func(ctx context.Context) any {
		return output.Seq(func(yield func(jsondoc.JSONDoc, error) bool) {
			r.sendRequest(ctx, client, opts, yield)
		})
	}
}

func (r *Runner) sendRequest(ctx context.Context, client *apiv2.Client, opts apiv2.RequestOptions, yield func(jsondoc.JSONDoc, error) bool) {
	common, err := r.Config.GetOutputCommonOptions(r.Cmd)
	if err != nil {
		yield(jsondoc.Empty(), err)
		return
	}
	paging := strings.EqualFold(opts.Method, http.MethodGet) && (common.IncludeAll || common.TotalPages > 1)
	// The page size of the requests: a request without pageSize (e.g. c8y api)
	// gets the server's default pages, so a short page does not end the paging.
	pageSize := requestPageSize(opts.Query)
	whole := r.Config.RawOutput() || common.WithTotalPages || common.WithTotalElements

	for page := int64(1); ; page++ {
		res := client.SendRequest(ctx, opts)
		if res.Error != nil {
			yield(jsondoc.Empty(), res.Error)
			return
		}
		body := res.Body()
		if res.Response == nil || len(body) == 0 {
			// A dry run, or a response without a body (e.g. 204)
			return
		}
		if r.Config.IsResponseOutput() || !isJSONDocument(body) {
			if err := r.writeResponseBody(res, body); err != nil {
				yield(jsondoc.Empty(), err)
			}
			return
		}

		property := r.resultProperty(body)
		count := 0
		docs := []jsondoc.JSONDoc{jsondoc.New(body)}
		if property != "" || body[0] == '[' {
			items := gjson.ParseBytes(body)
			if property != "" {
				items = items.Get(property)
			}
			count = len(items.Array())
			if !whole {
				docs = docs[:0]
				for _, item := range items.Array() {
					docs = append(docs, jsondoc.New([]byte(item.Raw)))
				}
			}
		}
		for _, doc := range docs {
			if !yield(doc, nil) {
				return
			}
		}

		// Follow the next page link of a collection
		next := gjson.GetBytes(body, "next").String()
		if !paging || property == "" || next == "" || count == 0 || count < pageSize {
			return
		}
		if common.TotalPages > 0 && page >= common.TotalPages {
			return
		}
		nextURL, err := url.Parse(next)
		if err != nil {
			return
		}
		if delay := r.Config.GetIncludeAllDelay(); delay > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(delay) * time.Millisecond):
			}
		}
		opts = apiv2.RequestOptions{
			Method:         http.MethodGet,
			Host:           opts.Host,
			Path:           nextURL.Path,
			Query:          nextURL.RawQuery,
			Accept:         opts.Accept,
			Headers:        opts.Headers,
			ProcessingMode: opts.ProcessingMode,
			IgnoreAccept:   opts.IgnoreAccept,
		}
	}
}

// requestPageSize returns the pageSize of a request's query (url.Values or an
// encoded string), or 0 when it has none.
func requestPageSize(query any) int {
	var values url.Values
	switch q := query.(type) {
	case url.Values:
		values = q
	case string:
		values, _ = url.ParseQuery(q)
	}
	size, _ := strconv.Atoi(values.Get(flags.FlagPageSize))
	return size
}

// resultProperty returns the array property which holds a response's items:
// the command's collection property (flags.WithCollectionProperty) when the
// response has it ("-" for none), or else the detected one (see
// collectionProperty).
func (r *Runner) resultProperty(body []byte) string {
	switch property := flags.GetCollectionPropertyFromAnnotation(r.Cmd); property {
	case "":
	case "-":
		return ""
	default:
		if gjson.GetBytes(body, property).IsArray() {
			return property
		}
	}
	return collectionProperty(body)
}

// isJSONDocument reports whether body is a JSON object or array.
func isJSONDocument(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return false
	}
	return gjson.ValidBytes(trimmed)
}

// collectionProperty returns the array property which holds a response's
// items, or "" when the response is not a collection: it has an id, several
// array properties, or more properties than a collection (e.g. the tenant
// statistics summary).
func collectionProperty(body []byte) string {
	doc := gjson.ParseBytes(body)
	if !doc.IsObject() || doc.Get("id").Exists() {
		return ""
	}
	property := ""
	arrays, keys := 0, 0
	doc.ForEach(func(key, value gjson.Result) bool {
		keys++
		if value.IsArray() {
			arrays++
			property = key.String()
		}
		return true
	})
	if arrays != 1 || keys > 10 {
		return ""
	}
	return property
}

// writeResponseBody writes a response body as is to --outputFileRaw, or else
// to stdout. The --outputFileRaw template can use {id} (the first id in the
// request path, or the response's id), {name}, {type}, {owner}, {filename}
// (from the Content-Disposition header) and {basename} (of the request path).
func (r *Runner) writeResponseBody(res *apiv2.RequestResult, body []byte) error {
	if path := r.Config.GetOutputFileRaw(); path != "" {
		_, err := fileutilities.WriteToFile(bytes.NewReader(body), path, false, isJSONDocument(body), responseFields(res, body))
		return err
	}
	out := r.Cmd.OutOrStdout()
	raw := r.Config.IsResponseOutput() || res.Header("Content-Type") == "application/octet-stream"
	if body = responseText(body, raw); len(body) == 0 {
		return nil
	}
	if _, err := out.Write(body); err != nil {
		return err
	}
	if r.Factory.IOStreams.IsStdoutTTY() && !bytes.HasSuffix(body, []byte("\n")) {
		_, err := fmt.Fprintln(out)
		return err
	}
	return nil
}

// responseText returns the bytes which are written for a response body which is
// not JSON. Only a binary download (and --raw output) is written as received. Any
// other body is text: it is trimmed and ends with a newline, so that the
// responses of several requests are written line by line (as in v2, which also
// writes an empty line for a body of only whitespace).
func responseText(body []byte, raw bool) []byte {
	if raw || len(body) == 0 {
		return body
	}
	body = bytes.TrimSpace(body)
	return append(body[:len(body):len(body)], '\n')
}

func responseFields(res *apiv2.RequestResult, body []byte) map[string]string {
	fields := map[string]string{}
	for _, prop := range []string{"id", "name", "type", "owner"} {
		fields[prop] = gjson.GetBytes(body, prop).String()
	}
	if _, params, err := mime.ParseMediaType(res.Header("Content-Disposition")); err == nil {
		if name, ok := params["filename"]; ok {
			fields["filename"] = name
		}
	}
	if req := res.Response.Request; req != nil {
		fields["basename"] = path.Base(req.URL.Path)
		for part := range strings.SplitSeq(req.URL.Path, "/") {
			if part != "" && isDigits(part) {
				fields["id"] = part
				break
			}
		}
	}
	return fields
}
