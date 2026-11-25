package cmdparser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/request"
	pkgurl "github.com/reubenmiller/go-c8y-cli/v2/pkg/url"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/types"
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
			return f.CreateModeEnabled(cmd)
		case "PUT":
			return f.UpdateModeEnabled(cmd)
		case "DELETE":
			return f.DeleteModeEnabled(cmd)
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

// Prepare adds the flag getters of the command's parameters (path, header,
// query and body) to the command options.
func (n *RuntimeCmd) Prepare(args []string) error {
	item := n.options.Spec
	subcmd := n.options
	factory := n.factory

	// Presets
	if subcmd.Spec.HasPreset() {
		var values *[]flags.GetOption
		switch subcmd.Spec.Preset.Type {
		case PresetGetIdentity:
			values = &subcmd.QueryParameter
		}
		if values != nil {
			for _, p := range subcmd.Spec.Preset.Extensions {
				*values = append(*values, GetOption(subcmd, &p, factory, args)...)
			}
		}
	}

	// path
	for _, p := range item.PathParameters {
		subcmd.Path.Options = append(subcmd.Path.Options, GetOption(subcmd, &p, factory, args)...)
	}
	subcmd.Path.Template = item.Path

	// header
	// Note: the custom headers (--header) and the processing mode are added by the runner
	for _, p := range item.HeaderParameters {
		subcmd.Header = append(subcmd.Header, GetOption(subcmd, &p, factory, args)...)
	}

	// query
	// Note: the custom query parameters (--customQueryParam) are added by the runner
	for _, p := range item.QueryParameters {
		subcmd.QueryParameter = append(subcmd.QueryParameter, GetOption(subcmd, &p, factory, args)...)

		// Support Cumulocity Query builder
		if len(p.Children) > 0 {
			queryOptions := []flags.GetOption{}
			for _, child := range p.Children {
				// Ignore special in-built values as these are handled separately
				if child.Name == "queryTemplate" || child.Name == "orderBy" {
					continue
				}
				queryOptions = append(queryOptions, GetOption(subcmd, &child, factory, args)...)
			}

			if subcmd.Spec.HasPreset() {
				switch subcmd.Spec.Preset.Type {
				case PresetQueryInventory:
					for _, p := range subcmd.Spec.Preset.Extensions {
						queryOptions = append(queryOptions, GetOption(subcmd, &p, factory, args)...)
					}
				case PresetQueryInventoryChildren:
					for _, p := range subcmd.Spec.Preset.Extensions {
						queryOptions = append(queryOptions, GetOption(subcmd, &p, factory, args)...)
					}
				}
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

	supportsFormData := false
	switch item.GetBodyContentType() {
	case "binary":
		subcmd.Body.IsBinary = true
		supportsFormData = true
	case "formdata":
		supportsFormData = true
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	case "jsonarray":
		subcmd.Body.DefaultValue = []byte("[]")
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	case "jsonobject":
		subcmd.Body.DefaultValue = []byte("{}")
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	default:
		subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDataFlagValue())
	}
	for _, p := range item.Body {
		switch p.Type {
		case "file", "attachment":
			subcmd.Body.UploadProgressSource = p.Name
			subcmd.FormData = append(subcmd.FormData, GetOption(subcmd, &p, factory, args)...)
		case "fileContents":
			subcmd.Body.UploadProgressSource = p.Name
			fallthrough
		default:
			if supportsFormData {
				subcmd.FormData = append(subcmd.FormData, GetOption(subcmd, &p, factory, args)...)
			} else {
				subcmd.Body.Options = append(subcmd.Body.Options, GetOption(subcmd, &p, factory, args)...)
			}
		}
	}

	subcmd.Body.Options = append(subcmd.Body.Options, cmdutil.WithTemplateValue(factory))
	subcmd.Body.Options = append(subcmd.Body.Options, flags.WithTemplateVariablesValue())

	for _, bodyTemplate := range item.BodyTemplates {
		if bodyTemplate.Type == "jsonnet" {
			if bodyTemplate.ApplyLast {
				subcmd.Body.Options = append(subcmd.Body.Options, flags.WithRequiredTemplateString(bodyTemplate.Template))
			} else {
				subcmd.Body.Options = append(subcmd.Body.Options, flags.WithDefaultTemplateString(bodyTemplate.Template))
			}
		}
	}

	return nil
}

// RunE executes the command on the c8ystream runner: the pipeline flag (or
// its values) drives the iteration, and each input item is sent as a request
// with the go-c8y v2 client. The runner handles the dry run, confirmation,
// workers and the output.
func (n *RuntimeCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	if err := ApplyCommandView(n.factory, cfg, n.options.Spec); err != nil {
		return err
	}

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.Filter(n.options.Spec.Filter...); err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("invalid filter in command '%s'. %s", n.options.Spec.Name, err))
	}

	// The pipeline flag drives the iteration (piped items, or the flag's own
	// values). The flag getters read the item's value from the driver, so
	// stdin is only read once.
	current := &currentValue{}
	pipeOpts := r.InputIterators.PipeOptions
	if pipeOpts != nil && pipeOpts.Name != "" {
		current.name = pipeOpts.Name
		if err := r.InputFlag(pipeOpts.Name); err != nil {
			if paramErr := new(flags.ParameterError); errors.As(err, &paramErr) {
				return cmderrors.NewUserError(err)
			}
			return err
		}
		pipeOpts.Source = current
	}

	client, err := r.Client()
	if err != nil {
		return err
	}
	refs := n.options.refs
	refs.client = client
	refs.tenant = n.factory.GetTenant()

	if err := n.Prepare(args); err != nil {
		return err
	}

	req, err := n.newRequestTemplate(cmd, r)
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		current.in = in
		refs.ctx = in.ResolveContext()
		defer func() { refs.ctx = nil }()

		opts, err := req.build(in)
		if err != nil {
			return nil, err
		}
		return r.RequestCall(client, opts), nil
	})
}

// requestTemplate holds the parts of a runtime command's request. Parts which
// depend on the input item are iterators, which are evaluated for each item.
type requestTemplate struct {
	method       string
	path         *flags.StringTemplate
	query        *flags.QueryTemplate
	headers      []header
	body         *mapbuilder.MapBuilder
	isBinary     bool
	formData     map[string]io.Reader
	accept       string
	contentType  string
	ignoreAccept bool
	progress     func(*http.Request) (*http.Request, error)
}

type header struct {
	name  string
	value any
}

// newRequestTemplate evaluates the command's flag getters
func (n *RuntimeCmd) newRequestTemplate(cmd *cobra.Command, r *c8ystream.Runner) (*requestTemplate, error) {
	cfg := r.Config
	spec := n.options.Spec
	inputIterators := r.InputIterators

	req := &requestTemplate{
		method:       spec.Method,
		accept:       spec.Accept,
		contentType:  spec.ContentType,
		ignoreAccept: cfg.IgnoreAcceptHeader(),
		isBinary:     n.options.Body.IsBinary,
	}

	// query parameters
	req.query = flags.NewQueryTemplate()
	if err := flags.WithQueryParameters(cmd, req.query, inputIterators, n.options.QueryParameter...); err != nil {
		return nil, cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return nil, cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParametersWithMapping(req.query, spec.FlagMapping)

	// headers
	for _, opt := range n.options.Header {
		name, value, err := opt(cmd, inputIterators)
		if err != nil {
			return nil, cmderrors.NewUserError(err)
		}
		if name != "" && value != nil {
			req.headers = append(req.headers, header{name: name, value: value})
		}
	}

	// form data
	req.formData = make(map[string]io.Reader)
	if err := flags.WithFormDataOptions(cmd, req.formData, inputIterators, n.options.FormData...); err != nil {
		return nil, cmderrors.NewUserError(err)
	}
	if err := flags.MakeFormDataReopenable(req.formData); err != nil {
		return nil, cmderrors.NewUserError(err)
	}

	// body
	if len(n.options.Body.DefaultValue) > 0 {
		req.body = mapbuilder.NewMapBuilderWithInit(n.options.Body.DefaultValue)
	} else {
		req.body = mapbuilder.NewInitializedMapBuilder(n.options.Body.Initialize)
	}
	if err := flags.WithBody(cmd, req.body, inputIterators, n.options.Body.Options...); err != nil {
		return nil, cmderrors.NewUserError(err)
	}

	// path parameters
	req.path = flags.NewStringTemplate(n.options.Path.Template)
	if err := flags.WithPathParameters(cmd, req.path, inputIterators, n.options.Path.Options...); err != nil {
		return nil, err
	}

	// upload progress bar
	if source := n.options.Body.UploadProgressSource; source != "" {
		req.progress = c8ybinary.AddProgress(
			cmd,
			source,
			cfg.GetProgressBar(n.factory.IOStreams.ErrOut, n.factory.IOStreams.IsStderrTTY()),
		)
	}
	return req, nil
}

// build returns the request of one input item
func (t *requestTemplate) build(in *c8ystream.Resolver) (apiv2.RequestOptions, error) {
	opts := apiv2.RequestOptions{
		Method:       t.method,
		Accept:       t.accept,
		ContentType:  t.contentType,
		IgnoreAccept: t.ignoreAccept,
	}

	// path (and any query parameters or host in the path)
	pathValue, _, err := t.path.GetNext()
	if err != nil {
		return opts, err
	}
	if hasUnresolvedVariables(pathValue) {
		return opts, cmderrors.ErrNoMatchesFound
	}
	query := url.Values{}
	u, err := parseURL(pkgurl.PathEscapePreserveQueryParameters(string(pathValue)))
	if err != nil {
		return opts, cmderrors.NewUserError(fmt.Sprintf("invalid path. %s", err))
	}
	if u.Host != "" {
		opts.Host = u.Scheme + "://" + u.Host
	}
	opts.Path = u.EscapedPath()
	if !strings.HasPrefix(opts.Path, "/") {
		opts.Path = "/" + opts.Path
	}
	for key, values := range u.Query() {
		query[key] = append(query[key], values...)
	}

	// query parameters. The flag getters pre-encode some values (for the raw
	// v1 query), while the v2 client encodes the values itself, so they are
	// decoded once.
	values, _, err := t.query.Execute(false)
	if err != nil {
		return opts, err
	}
	for key, items := range values {
		for _, item := range items {
			if decoded, err := url.QueryUnescape(item); err == nil {
				item = decoded
			}
			query.Add(key, item)
		}
	}
	opts.Query = query

	// headers
	for _, h := range t.headers {
		value, err := headerValue(h.value)
		if err != nil {
			return opts, err
		}
		switch {
		case value == "":
		case strings.EqualFold(h.name, "Content-Type"):
			opts.ContentType = value
		case strings.EqualFold(h.name, "Accept"):
			opts.Accept = value
		default:
			if opts.Headers == nil {
				opts.Headers = map[string]string{}
			}
			opts.Headers[h.name] = value
		}
	}

	if !request.RequestSupportsBody(t.method) {
		return opts, nil
	}

	// form data (the readers are recreated for each request). The file
	// part is named by the form's filename value (e.g. --name), as in v1.
	if len(t.formData) > 0 {
		formData, err := flags.ReopenFormData(t.formData)
		if err != nil {
			return opts, cmderrors.NewUserError(err)
		}
		opts.FormData = c8ystream.FormParts(formData)
	}

	// body
	switch {
	case t.body.HasFile():
		file, err := t.body.GetFile()
		if err != nil {
			return opts, cmderrors.NewUserError(err)
		}
		opts.Body = file
	case t.isBinary:
		// a binary body is only sent from a file
	case t.body.HasRaw():
		opts.Body = []byte(t.body.GetRaw())
	default:
		var contents []byte
		if input := in.Input(); input != nil {
			contents, err = t.body.MarshalJSONWithInput(input)
		} else {
			contents, err = t.body.MarshalJSON()
		}
		if err != nil {
			return opts, cmderrors.NewUserError(err)
		}
		if len(contents) > 0 {
			// the template output is indented; send it compact, as v1 did
			var compact bytes.Buffer
			if json.Compact(&compact, contents) == nil {
				contents = compact.Bytes()
			}
			opts.Body = contents
			if opts.ContentType == "" {
				opts.ContentType = types.MimeTypeApplicationJSON
			}
		}
	}

	if t.progress != nil && (opts.Body != nil || len(opts.FormData) > 0) {
		progress := t.progress
		opts.PrepareRequest = func(req *http.Request) error {
			out, err := progress(req)
			if err == nil && out != nil {
				req.Body = out.Body
			}
			return err
		}
	}
	return opts, nil
}

// headerValue returns the value of a header for the current item
func headerValue(value any) (string, error) {
	switch v := value.(type) {
	case iterator.Iterator:
		b, _, err := v.GetNext()
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return string(b), nil
	case []string:
		return strings.Join(v, ","), nil
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// currentValue is the pipeline source of the flag getters: the value of the
// pipeline flag for the item being built
type currentValue struct {
	name string
	in   *c8ystream.Resolver
}

func (c *currentValue) GetNext() ([]byte, interface{}, error) {
	if c.in == nil {
		return nil, nil, io.EOF
	}
	return []byte(c.in.String(c.name)), c.in.Input(), nil
}

func (c *currentValue) IsBound() bool { return true }

func (c *currentValue) MarshalJSON() ([]byte, error) { return iterator.MarshalJSON(c) }

func hasUnresolvedVariables(v []byte) bool {
	left := bytes.Index(v, []byte("{"))
	right := bytes.Index(v, []byte("}"))
	return left > -1 && right > -1 && right > left
}

func parseURL(s string) (*url.URL, error) {
	// Default to https if a scheme is missing
	rawURL := s
	if p, err := url.Parse(s); err == nil && p.Host == "" && p.Scheme != "" {
		rawURL = "https://" + s
	}
	return url.Parse(rawURL)
}
