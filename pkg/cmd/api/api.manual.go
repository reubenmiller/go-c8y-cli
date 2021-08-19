package api

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/pkg/request"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdAPI struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	method         string
	flagHost       string
	keepProperties bool
}

var allowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

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

			$ c8y api GET "/alarm/alarms?pageSize=10&status=ACTIVE"
			Get a list of alarms with custom query parameters

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
	cmd.Flags().BoolVar(&ccmd.keepProperties, "keepProperties", true, "Don't strip Cumulocity properties from the data property, i.e. source etc.")
	cmd.Flags().StringVar(&ccmd.flagHost, "host", "", "host to use for the rest request. If empty, then the session's host will be used")

	flags.WithOptions(
		cmd,
		flags.WithData(),
		flags.WithTemplate(),
		// Include device management url links (configuration dump, firmware and software)
		flags.WithExtendedPipelineSupport("url", "url", false, "url", "c8y_Firmware.url", "c8y_Software.url", "self", "responseSelf"),
	)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("method", "GET", "POST", "DELETE", "PUT"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdAPI) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	method := n.method

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithProcessingModeValue(),
		flags.WithStringValue("accept", "Accept"),
		flags.WithStringValue("contentType", "Content-Type"),
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	var uri string
	if len(args) == 1 {
		isMethod := false
		for _, m := range allowedMethods {
			if strings.EqualFold(m, args[0]) {
				isMethod = true
				break
			}
		}
		if isMethod {
			method = args[0]
		} else {
			uri = args[0]
		}
	} else if len(args) > 1 {
		method = args[0]
		uri = args[1]
	}

	// path parameters
	urlTemplate := "{url}"
	if uri != "" {
		urlTemplate = uri
	}
	if cmd.Flags().Changed("url") {
		if v, err := cmd.Flags().GetString("url"); err == nil {
			urlTemplate = v
		}
	}
	if !strings.Contains(urlTemplate, "{url}") && strings.Contains(urlTemplate, "%s") {
		urlTemplate = strings.Replace(urlTemplate, "%s", "{url}", -1)
	}

	path := flags.NewStringTemplate(urlTemplate)

	// set a default uri to prevent unresolved template variables when
	// stdin is not being used
	path.SetVariable("url", uri)
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("url", "url"),
	)

	// path.
	if err != nil {
		cfg.Logger.Warn("something is not being detected")
		return err
	}

	method = strings.ToUpper(method)

	if !(method == "GET" || method == "POST" || method == "PUT" || method == "DELETE") {
		return cmderrors.NewUserError("Invalid method. Only GET, PUT, POST and DELETE are accepted")
	}

	if method == "PUT" {
		if err := n.factory.UpdateModeEnabled(); err != nil {
			return err
		}
	}
	if method == "POST" {
		if err := n.factory.CreateModeEnabled(); err != nil {
			return err
		}
	}
	if method == "DELETE" {
		if err := n.factory.DeleteModeEnabled(); err != nil {
			return err
		}
	}

	baseURL, _ := url.Parse(uri)

	// query parameters
	query := flags.NewQueryTemplate()
	for key, values := range baseURL.Query() {
		query.SetVariable(key, values)
	}

	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
	)
	if err != nil {
		return nil
	}
	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	var host string
	if n.flagHost != "" {
		host = n.flagHost
	}

	req := c8y.RequestOptions{
		Method:       method,
		Host:         host,
		Path:         baseURL.Path,
		Query:        queryValue,
		Header:       headers,
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		ResponseData: nil,
	}

	if method == "PUT" || method == "POST" {
		body := mapbuilder.NewInitializedMapBuilder()
		err = flags.WithBody(
			cmd,
			body,
			inputIterators,
			flags.WithDataValueAdvanced(!n.keepProperties, !request.HasJSONHeader(&headers), flags.FlagDataName, ""),
			cmdutil.WithTemplateValue(cfg),
			flags.WithTemplateVariablesValue(),
		)

		if err != nil {
			return cmderrors.NewUserError(err)
		}

		if cmd.Flags().Changed("template") || cmd.Flags().Changed("data") {
			req.Body = body
		}

		// get file info
		// form data
		formData := make(map[string]io.Reader)
		err = flags.WithFormDataOptions(
			cmd,
			formData,
			inputIterators,
			flags.WithFormDataFileAndInfo("file", "formdata")...,
		)
		if err != nil {
			return cmderrors.NewUserError(err)
		}
		req.FormData = formData
	}

	// Hide usage for system errors
	cmd.SilenceUsage = true

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
