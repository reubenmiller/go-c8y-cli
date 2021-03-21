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
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdAPI struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	flagHost string
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
### Get a list of alarms
$ c8y api GET /alarm/alarms

### Get a list of alarms with custom query parameters
c8y api GET "/alarm/alarms?pageSize=10&status=ACTIVE"

### Create a new alarm
c8y api POST "alarm/alarms" --data "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source={'id': '12345'}"
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().String("file", "", "File to be uploaded as a binary")
	cmd.Flags().StringSliceP("header", "H", nil, "headers. i.e. --header \"Accept: value\"")
	cmd.Flags().String("accept", "", "accept (header)")
	cmd.Flags().String("contentType", "", "content type (header)")
	cmd.Flags().StringVar(&ccmd.flagHost, "host", "", "host to use for the rest request. If empty, then the session's host will be used")

	flags.WithOptions(
		cmd,
		flags.WithData(),
		flags.WithTemplate(),
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
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	method := "get"

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithProcessingModeValue(),
		flags.WithStringValue("accept", "Accept"),
		flags.WithStringValue("contentType", "Content-Type"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	if values, err := cmd.Flags().GetStringSlice("header"); err == nil && len(values) > 0 {
		for _, v := range values {
			parts := strings.SplitN(v, ":", 2)
			if len(parts) != 2 {
				log.Warningf("Invalid header. %s", v)
				continue
			}
			log.Debugf("Setting header: name=%s, value=%s", parts[0], parts[1])
			headers.Add(parts[0], strings.TrimSpace(parts[1]))
		}
	}

	var uri string
	if len(args) == 1 {
		uri = args[0]
	} else if len(args) > 1 {
		method = args[0]
		uri = args[1]
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

	var host string
	if n.flagHost != "" {
		host = n.flagHost
	}

	req := c8y.RequestOptions{
		Method:       method,
		Host:         host,
		Path:         baseURL.Path,
		Query:        baseURL.RawQuery,
		Header:       headers,
		DryRun:       cfg.DryRun(),
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		ResponseData: nil,
	}

	if method == "PUT" || method == "POST" {
		body := mapbuilder.NewInitializedMapBuilder()
		err = flags.WithBody(
			cmd,
			body,
			inputIterators,
			flags.WithDataFlagValue(),
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
			flags.WithFormDataFileAndInfo("file", "data")...,
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
