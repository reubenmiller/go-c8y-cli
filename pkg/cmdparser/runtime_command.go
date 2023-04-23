package cmdparser

import (
	"fmt"
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type RuntimeCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	options *CmdOptions
}

// NewRuntimeCmd creates a command which is created at runtime
func NewRuntimeCmd(f *cmdutil.Factory, options *CmdOptions) *RuntimeCmd {
	ccmd := &RuntimeCmd{
		factory: f,
		options: options,
	}
	cmd := options.Command
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Mode checks
		switch options.Spec.GetMethod() {
		case "POST":
			return f.CreateModeEnabled()
		case "PUT":
			return f.UpdateModeEnabled()
		case "DELETE":
			return f.DeleteModeEnabled()
		}
		return nil
	}
	cmd.RunE = ccmd.RunE
	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		options.Completion...,
	)

	flags.WithOptions(
		cmd,
		options.Runtime...,
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)
	return ccmd
}

func (n *RuntimeCmd) Prepare(args []string) error {
	item := n.options.Spec
	subcmd := n.options
	factory := n.factory

	client, err := factory.Client()
	if err != nil {
		return err
	}
	cfg, err := factory.Config()
	if err != nil {
		return err
	}

	// path
	for _, p := range item.PathParameters {
		subcmd.Path.Options = append(subcmd.Path.Options, GetOption(subcmd, &p, factory, cfg, client, args)...)
	}
	subcmd.Path.Template = item.Path

	// header
	subcmd.Header = append(subcmd.Header, flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"))
	for _, p := range item.HeaderParameters {
		subcmd.Header = append(subcmd.Header, GetOption(subcmd, &p, factory, cfg, client, args)...)
	}

	if item.SupportsProcessingMode() {
		subcmd.Header = append(subcmd.Header, flags.WithProcessingModeValue())
	}

	// query
	subcmd.QueryParameter = append(subcmd.QueryParameter, flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"))

	for _, p := range item.QueryParameters {
		subcmd.QueryParameter = append(subcmd.QueryParameter, GetOption(subcmd, &p, factory, cfg, client, args)...)

		// Support Cumulocity Query builder
		if len(p.Children) > 0 {
			queryOptions := []flags.GetOption{}
			for _, child := range p.Children {
				// Ignore special in-built values as these are handled separately
				if child.Name == "queryTemplate" || child.Name == "orderBy" {
					continue
				}
				queryOptions = append(queryOptions, GetOption(subcmd, &child, factory, cfg, client, args)...)
			}
			subcmd.QueryParameter = append(subcmd.QueryParameter, flags.WithCumulocityQuery(queryOptions, p.GetTargetProperty()))
		}
	}

	// body
	requiredBodyKeys := []string{}
	requiredBodyKeys = append(requiredBodyKeys, item.BodyRequiredKeys...)
	for _, p := range item.Body {
		if p.IsRequired() {
			requiredBodyKeys = append(requiredBodyKeys, p.GetTargetProperty())
		}
	}
	if len(requiredBodyKeys) > 0 {
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithRequiredProperties(requiredBodyKeys...))
	}

	if len(item.Body) > 0 {
		if item.Method == "PUT" || item.Method == "POST" {
			subcmd.Body.Initialize = true
		}
	}

	switch item.GetBodyContentType() {
	case "binary":
		subcmd.Body.IsBinary = true
	case "formdata":
	default:
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	}
	for _, p := range item.Body {
		subcmd.Body.Options = append(subcmd.Body.Options, GetOption(subcmd, &p, factory, cfg, client, args)...)
	}

	subcmd.Body.Options = append(subcmd.Body.Options, cmdutil.WithTemplateValue(factory, cfg))
	subcmd.Body.Options = append(subcmd.Body.Options, flags.WithTemplateVariablesValue())

	for _, bodyTemplate := range item.BodyTemplates {
		// TODO: Check if other body templates should be supported or not
		if bodyTemplate.Type == "jsonnet" {
			subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDefaultTemplateString(bodyTemplate.Template))
		}
	}

	return nil
}

// RunE executes the command
func (n *RuntimeCmd) RunE(cmd *cobra.Command, args []string) error {

	if err := n.Prepare(args); err != nil {
		return err
	}

	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		n.options.QueryParameter...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		n.options.Header...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
		n.options.FormData...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(n.options.Body.Initialize)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		n.options.Body.Options...,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate(n.options.Path.Template)
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		n.options.Path.Options...,
	)
	if err != nil {
		return err
	}

	var req *c8y.RequestOptions
	if n.options.Body.IsBinary {
		req = &c8y.RequestOptions{
			Method:       n.options.Spec.Method,
			Path:         path.GetTemplate(),
			Query:        queryValue,
			Body:         body.GetFileContents(),
			FormData:     formData,
			Header:       headers,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
			DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		}
	} else {
		req = &c8y.RequestOptions{
			Method:       n.options.Spec.Method,
			Path:         path.GetTemplate(),
			Query:        queryValue,
			Body:         body,
			FormData:     formData,
			Header:       headers,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
			DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		}
	}

	return n.factory.RunWithWorkers(client, cmd, req, inputIterators)
}
