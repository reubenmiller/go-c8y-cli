package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/requestutil"
	_url "github.com/reubenmiller/go-c8y-cli/v2/pkg/url"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/types"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type CmdAPI struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	method         string
	flagHost       string
	keepProperties bool
}

var allowedMethods = []string{
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
	http.MethodOptions,
}

func NewSubCommand(f *cmdutil.Factory) *CmdAPI {
	ccmd := &CmdAPI{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:     "api",
		Aliases: []string{"rest"},
		Short:   "Send api request",
		Long:    `Send an authenticated api request to a given endpoint`,
		Example: heredoc.Doc(`
			$ c8y api GET /alarm/alarms
			Get a list of alarms

			$ c8y api GET "/alarm/alarms&status=ACTIVE" --pageSize 10
			Get a list of alarms with custom query parameters

			$ c8y api GET "/dosomething" --host localhost:8080
			Send a GET request to a custom hosted service running on localhost port 8080. You can force https or http by adding them as a prefix to the host.

			$ c8y api GET "/dosomething" --host $C8Y_HOST/service/mymicroservice --dry
			Send a request to a microservice hosted in the cloud. The /dosomething will be appended to the /service/mymicroservice path provided by the host flag.
			Results in the following request: GET /service/mymicroservice/dosomething sent to the  current session URL.

			$ c8y api GET "/alarm/alarms&status=ACTIVE" --pageSize 1 --withTotalPages
			Get a total ACTIVE alarms

			$ c8y api POST "alarm/alarms" --data "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source.id='12345'" --keepProperties
			Create a new alarm

			$ c8y activitylog list --filter "method like POST" | c8y api --method DELETE
			Get items created via POST from the activity log and delete them 

			$ echo -e "/inventory/1111\n/inventory/2222" | c8y api --method PUT --template "{myScript: {lastUpdated: _.Now() }}"
			Pipe a list of urls and execute HTTP PUT and use a template to generate the body

			$ echo "12345" | c8y api PUT "/service/example" --template "{id: input.value}"
			Send a PUT request to a fixed url, but use the piped input to build the request's body

			$ echo "12345" | c8y api PUT "/service/example/%s" --template "{id: input.value}"
			Send a PUT request using the piped input in both the url and the request's body ('%s' will be replaced with the current piped input line)

			$ echo "{\"url\": \"/service/custom/endpoint\",\"body\":{\"name\": \"peter\"}}" | c8y api POST --template "input.value.body"
			Build a custom request using piped json. The input url property will be mapped to the --url flag, and use
			a template to also build the request's body from the piped input data.

			$ c8y api POST /example -F file=@./mydevice01.pem -F parentAssetId=1111 -F assetTypeId=2222
			Send a multipart form-data request with a file and custom form fields (similar to curl's -F option)

			$ c8y api POST /example --form "file=@./data.bin;type=application/octet-stream;filename=custom.bin" --form "description=<./notes.txt"
			Upload a file with a custom content type and filename, and set a form field to the contents of a file
		`),
		Args: cobra.MaximumNArgs(2),
		RunE: ccmd.RunE,
	}

	cmd.Flags().String("url", "", "URL path. Any reference to '%s' will be replaced with the current input value (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.method, "method", "GET", "HTTP method")
	cmd.Flags().String("file", "", "File to be uploaded as a binary")
	cmd.Flags().String("accept", "", "accept (header)")
	cmd.Flags().String("contentType", "", "content type (header)")
	cmd.Flags().String("formdata", "", "form data (json or shorthand json)")
	cmd.Flags().StringArrayP("form", "F", nil, "Multipart form field (curl style). Use name=value for text, name=@path[;type=mime/type][;filename=name] for a file, or name=<path for text read from a file. Can be used multiple times. Implies POST if no method is given")
	cmd.Flags().BoolVar(&ccmd.keepProperties, "keepProperties", true, "Don't strip Cumulocity properties from the data property, i.e. source etc.")
	cmd.Flags().StringVar(&ccmd.flagHost, "host", "", "host to use for the rest request. If empty, then the session's host will be used")

	flags.WithOptions(
		cmd,
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		// Include device management url links (configuration dump, firmware and software)
		flags.WithExtendedPipelineSupport("url", "url", false, "url", "c8y_Firmware.url", "c8y_Software.url", "self", "responseSelf"),
	)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("method", allowedMethods...),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdAPI) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	method := n.method
	var uri string
	if len(args) == 1 {
		if isMethod(args[0]) {
			method = args[0]
		} else {
			uri = args[0]
		}
	} else if len(args) > 1 {
		method = args[0]
		uri = args[1]
	}

	// Use POST when sending form fields and no method was given (same as curl)
	if cmd.Flags().Changed("form") && !cmd.Flags().Changed("method") && !(len(args) > 0 && isMethod(args[0])) {
		method = http.MethodPost
	}
	method = strings.ToUpper(method)
	if !isMethod(method) {
		return cmderrors.NewUserError(fmt.Sprintf("Invalid method. Only %s are accepted", strings.Join(allowedMethods, ", ")))
	}

	switch method {
	case http.MethodPut:
		err = n.factory.UpdateModeEnabled(cmd)
	case http.MethodPost:
		err = n.factory.CreateModeEnabled(cmd)
	case http.MethodDelete:
		err = n.factory.DeleteModeEnabled(cmd)
	}
	if err != nil {
		return err
	}

	// --host sets the scheme and host, and its path prefixes the url
	var host, pathPrefix string
	if n.flagHost != "" {
		u, err := _url.ParseURL(n.flagHost)
		if err != nil {
			return err
		}
		host = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		pathPrefix = u.Path
	}

	// The url template: '{url}' (or '%s') is replaced with the url of each
	// input item (piped value or --url)
	urlTemplate := "{url}"
	if uri != "" {
		urlTemplate = uri
	}
	// An explicit --url template takes each item's url from the piped item
	urlFlagTemplate := false
	if cmd.Flags().Changed("url") {
		if v, err := cmd.Flags().GetString("url"); err == nil && !flags.IsRuntimePipelineProperty(v) {
			if _, isRef := iterator.InputReference(v); !isRef {
				urlTemplate = v
				urlFlagTemplate = true
			}
		}
	}
	if !strings.Contains(urlTemplate, "{url}") {
		urlTemplate = strings.ReplaceAll(urlTemplate, "%s", "{url}")
	}

	if cmd.Flags().Changed("form") {
		if !requestutil.RequestSupportsBody(method) {
			return cmderrors.NewUserError(fmt.Sprintf("--form can not be used with a %s request", method))
		}
		if cmd.Flags().Changed("template") || cmd.Flags().Changed("data") {
			return cmderrors.NewUserError("--form can not be used together with --data or --template. Form data replaces the request body")
		}
	}

	accept, _ := cmd.Flags().GetString("accept")
	contentType, _ := cmd.Flags().GetString("contentType")

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	// The item's url comes from the pipe or --url, unless --url is the
	// template: then each piped item gives the url
	itemURL := func(in *c8ystream.Resolver) string { return in.String("url") }
	if urlFlagTemplate {
		if err := r.Input(); err != nil {
			return err
		}
		pipeOpts, err := flags.GetPipeOptionsFromAnnotation(cmd)
		if err != nil {
			return err
		}
		itemURL = func(in *c8ystream.Resolver) string { return pipedURL(in.Input(), pipeOpts) }
	} else if err := r.InputFlag("url"); err != nil {
		return err
	}

	hasBody := requestutil.RequestSupportsBody(method) && (cmd.Flags().Changed("template") || cmd.Flags().Changed("data"))
	if hasBody {
		headers := http.Header{}
		if contentType != "" {
			headers.Set("Content-Type", contentType)
		}
		if err := r.Body(
			flags.WithDataValueAdvanced(!n.keepProperties, !requestutil.HasJSONHeader(&headers), flags.FlagDataName, ""),
			cmdutil.WithTemplateValue(n.factory),
			flags.WithTemplateVariablesValue(),
		); err != nil {
			return err
		}
	}

	// Form data (--file, --formdata and --form) is sent with every request
	var form map[string]io.Reader
	if requestutil.RequestSupportsBody(method) {
		form = make(map[string]io.Reader)
		if err := flags.WithFormDataOptions(cmd, form, r.InputIterators, flags.WithFormDataFileAndInfo("file", "formdata")...); err != nil {
			return cmderrors.NewUserError(err)
		}
		if err := flags.MakeFormDataReopenable(form); err != nil {
			return cmderrors.NewUserError(err)
		}
		if err := flags.WithFormFields(cmd, form, "form"); err != nil {
			return cmderrors.NewUserError(err)
		}
	}

	var paging url.Values
	if method == http.MethodGet {
		common, err := cfg.GetOutputCommonOptions(cmd)
		if err != nil {
			return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
		}
		paging = pagingParameters(common)
	}

	progressBar := cfg.GetProgressBar(n.factory.IOStreams.ErrOut, n.factory.IOStreams.IsStderrTTY())
	addProgress := c8ybinary.AddProgress(cmd, "file", progressBar)

	client, err := r.Client()
	if err != nil {
		return err
	}

	// Hide usage for system errors
	cmd.SilenceUsage = true

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		urlHost, path, query, err := requestURL(urlTemplate, itemURL(in), pathPrefix)
		if err != nil {
			return nil, cmderrors.NewUserError(err)
		}
		requestHost := host
		if urlHost != "" {
			// an absolute url is requested from its own host (v2.59.0)
			if log, logErr := n.factory.Logger(); logErr == nil && log != nil {
				log.Warnf("Parsing url in request. %s", strings.TrimPrefix(strings.TrimPrefix(urlHost, "https://"), "http://"))
			}
			requestHost = urlHost
		}
		// The paging flags and the default page size replace the paging
		// parameters of the url (v2.59.0: "?pageSize=7" is sent as
		// "pageSize=20", or the value of --pageSize)
		for key, values := range paging {
			query[key] = values
		}

		opts := apiv2.RequestOptions{
			Method:       method,
			Host:         requestHost,
			Path:         path,
			Query:        query,
			Accept:       accept,
			ContentType:  contentType,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
		}
		if hasBody {
			if raw, ok := in.RawBody(); ok {
				opts.Body = raw
			} else {
				body, err := in.Body()
				if err != nil {
					return nil, cmderrors.NewUserError(err)
				}
				opts.Body = []byte(body)
				if opts.ContentType == "" {
					opts.ContentType = types.MimeTypeApplicationJSON
				}
			}
		}
		if len(form) > 0 {
			formData, err := flags.ReopenFormData(form)
			if err != nil {
				return nil, cmderrors.NewUserError(err)
			}
			opts.FormData = c8ystream.FormParts(formData)
			opts.PrepareRequest = addProgress
		}
		return r.RequestCall(client, opts), nil
	})
}

func isMethod(v string) bool {
	for _, m := range allowedMethods {
		if strings.EqualFold(m, v) {
			return true
		}
	}
	return false
}

// pipedURL returns the url of a piped item: a JSON object's pipeline property
// (or the first of its aliases it has), or else the whole line.
func pipedURL(input []byte, opts *flags.PipelineOptions) string {
	raw := bytes.TrimSpace(input)
	if !jsonUtilities.IsJSONObject(raw) {
		return string(raw)
	}
	if opts == nil {
		return ""
	}
	for _, property := range append([]string{opts.Property}, opts.Aliases...) {
		if v := gjson.GetBytes(raw, property); property != "" && v.Exists() {
			return v.String()
		}
	}
	return ""
}

// requestURL returns the host (only for an absolute url), the path and the
// query of a request: the url template with '{url}' replaced by the item's
// url, prefixed with the --host path. An absolute url (an argument or a piped
// self link) is requested from its own host, as in v2.59.0; the --host path
// prefix does not apply to it.
func requestURL(template, itemURL, prefix string) (string, string, url.Values, error) {
	raw := strings.ReplaceAll(template, "{url}", itemURL)
	host := ""
	if u, err := url.Parse(raw); err == nil && u.IsAbs() && u.Host != "" {
		host = u.Scheme + "://" + u.Host
		raw = u.RequestURI()
		prefix = ""
	}
	path, rawQuery, _ := strings.Cut(raw, "?")
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", "", nil, fmt.Errorf("invalid query parameters. %w", err)
	}
	if prefix != "" {
		path = strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(path, "/")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return host, path, query, nil
}

// pagingParameters returns the query parameters of the paging flags
// (--pageSize, --currentPage, --withTotalPages and --withTotalElements).
func pagingParameters(common config.CommonCommandOptions) url.Values {
	query := url.Values{}
	if common.CurrentPage > 0 {
		query.Set(flags.FlagCurrentPage, fmt.Sprintf("%d", common.CurrentPage))
	}
	if common.PageSize > 0 {
		query.Set(flags.FlagPageSize, fmt.Sprintf("%d", common.PageSize))
	}
	if common.WithTotalPages {
		query.Set(flags.FlagWithTotalPages, "true")
	}
	if common.WithTotalElements {
		query.Set(flags.FlagWithTotalElements, "true")
	}
	return query
}
