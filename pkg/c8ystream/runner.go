package c8ystream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// Args holds the resolved arguments for one service call: the values the
// bound sources produced for a single input item. Unlike the v1
// RequestInputIterators (which iterated towards a raw HTTP request), Args is
// shaped for typed go-c8y service calls.
type Args struct {
	// Index is the zero-based position of this item in the input stream.
	Index int64
	// Input is the raw input item (e.g. the piped JSON line) that produced
	// these arguments, for logging/teeing.
	Input any
	// Body is the resolved body when a Body source is bound.
	Body json.RawMessage

	values map[string]string
	query  url.Values
}

// Value returns the resolved value of a named Param source (e.g. "id").
func (a Args) Value(name string) string {
	return a.values[name]
}

// Query returns the resolved query parameter for the given key when a Query
// source is bound. The value is returned decoded: the v1 flag formatters
// pre-encode values for the raw-query path, while v2 typed options are
// encoded by the client itself.
func (a Args) Query(key string) string {
	if a.query == nil {
		return ""
	}
	raw := a.query.Get(key)
	if v, err := url.QueryUnescape(raw); err == nil {
		return v
	}
	return raw
}

// Source is a per-item input source bound to a Runner. The command does not
// care how a source resolves its value (query expression, named parameter,
// body, ...) — only that all bound sources advance together, one step per
// input item, each contributing its resolved value to Args.
type Source interface {
	bind(r *Runner) error
	// advance resolves the source's value for the next input item into args.
	// input carries the item pulled by an earlier source so later sources
	// (e.g. a body) evaluate against the same item instead of advancing
	// their own iterators.
	advance(args *Args, input *any) error
	// iterates reports whether the source ends on its own via io.EOF
	// (e.g. flag-provided values), used for the run-once termination rule.
	iterates() bool
}

// Query declares a query source built from the given flag options
// (e.g. flags.WithCumulocityQuery). Resolved values are available per item
// via Args.Query.
func Query(opts ...flags.GetOption) Source {
	return &querySource{opts: opts}
}

// Param declares a named value source (e.g. "id"), resolved per input item
// via the path-parameter machinery including any name -> id lookup options.
// The resolved value is available via Args.Value(name).
func Param(name string, opts ...flags.GetOption) Source {
	return &paramSource{name: name, opts: opts}
}

// Body declares a body source built from the given flag options (data flag,
// templates, piped values). The body is evaluated once per input item and
// available via Args.Body. It is always advanced last so it can be evaluated
// against the input item pulled by the other sources.
func Body(opts ...flags.GetOption) Source {
	return &bodySource{opts: opts}
}

type querySource struct {
	opts []flags.GetOption
	tpl  *flags.QueryTemplate
}

func (s *querySource) bind(r *Runner) error {
	s.tpl = flags.NewQueryTemplate()
	if err := flags.WithQueryParameters(r.Cmd, s.tpl, r.InputIterators, s.opts...); err != nil {
		return cmderrors.NewUserError(err)
	}
	return nil
}

func (s *querySource) advance(args *Args, input *any) error {
	values, in, err := s.tpl.Execute(false)
	if err != nil {
		return err
	}
	args.query = values
	if *input == nil {
		*input = in
	}
	return nil
}

func (s *querySource) iterates() bool { return false }

type paramSource struct {
	name string
	opts []flags.GetOption
	tpl  *flags.StringTemplate
}

func (s *paramSource) bind(r *Runner) error {
	s.tpl = flags.NewStringTemplate("{" + s.name + "}")
	return flags.WithPathParameters(r.Cmd, s.tpl, r.InputIterators, s.opts...)
}

func (s *paramSource) advance(args *Args, input *any) error {
	raw, in, err := s.tpl.GetNext()
	if err != nil {
		return err
	}
	value := string(raw)
	if value == "" || strings.ContainsAny(value, "{}") {
		return fmt.Errorf("could not resolve %q: %w", s.name, cmderrors.ErrNoMatchesFound)
	}
	if args.values == nil {
		args.values = make(map[string]string)
	}
	args.values[s.name] = value
	if *input == nil {
		*input = in
	}
	return nil
}

// Param values come from a flag or the pipeline either way, so their
// iterators end on their own.
func (s *paramSource) iterates() bool { return true }

type bodySource struct {
	opts    []flags.GetOption
	builder *mapbuilder.MapBuilder
}

func (s *bodySource) bind(r *Runner) error {
	s.builder = mapbuilder.NewInitializedMapBuilder(true)
	if err := flags.WithBody(r.Cmd, s.builder, r.InputIterators, s.opts...); err != nil {
		return cmderrors.NewUserError(err)
	}
	return nil
}

func (s *bodySource) advance(args *Args, input *any) error {
	var (
		bodyBytes []byte
		err       error
	)
	if *input != nil {
		// Another source already pulled the input item: evaluate the body
		// against it rather than advancing the body's own iterators.
		bodyBytes, err = s.builder.MarshalJSONWithInput(*input)
	} else {
		bodyBytes, err = s.builder.MarshalJSON()
	}
	if err != nil {
		return err
	}
	args.Body = json.RawMessage(bodyBytes)
	if *input == nil {
		*input = bodyBytes
	}
	return nil
}

func (s *bodySource) iterates() bool { return false }

// Runner owns the command-agnostic lifecycle of a v2-based command: pipeline
// annotations, the consolidated input iterator, context construction
// (dry-run, processing mode), the worker pool, error collection, the standard
// output stages and rendering.
//
// A command binds the per-item sources it needs and provides a single
// callback that turns Args into a service call.
type Runner struct {
	Cmd            *cobra.Command
	Factory        *cmdutil.Factory
	Config         *config.Config
	InputIterators *flags.RequestInputIterators

	sources []Source
	index   int64
}

// NewRunner resolves the command configuration and pipeline annotations.
func NewRunner(cmd *cobra.Command, f *cmdutil.Factory) (*Runner, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return nil, err
	}
	return &Runner{
		Cmd:            cmd,
		Factory:        f,
		Config:         cfg,
		InputIterators: inputIterators,
	}, nil
}

// Bind attaches the given sources to the runner. Sources advance in the
// order given, except Body sources which always advance last (so they can be
// evaluated against the input item pulled by the other sources).
func (r *Runner) Bind(sources ...Source) error {
	ordered := make([]Source, 0, len(sources))
	var bodies []Source
	for _, s := range sources {
		if _, ok := s.(*bodySource); ok {
			bodies = append(bodies, s)
			continue
		}
		ordered = append(ordered, s)
	}
	ordered = append(ordered, bodies...)

	for _, s := range ordered {
		if err := s.bind(r); err != nil {
			return err
		}
	}
	r.sources = append(r.sources, ordered...)
	return nil
}

// Client builds a go-c8y v2 client from the session via the legacy factory
// client, which owns credential resolution (encrypted session decryption,
// passphrase prompting, token refresh), so both clients always agree.
// --verbose/--debug enable the v2 client's request/response logging.
//
// TODO: once the legacy client is retired, resolve credentials directly on
// the factory and carry over proxy/TLS/timeout settings.
func (r *Runner) Client() (*apiv2.Client, error) {
	legacy, err := r.Factory.Client()
	if err != nil {
		return nil, err
	}
	return apiv2.NewClient(apiv2.ClientOptions{
		BaseURL: legacy.BaseURL.String(),
		Auth: authentication.AuthOptions{
			Tenant:   legacy.TenantName,
			Username: legacy.Username,
			Password: legacy.Password,
			Token:    legacy.Token,
		},
		Debug:     r.Config.Verbose() || r.Config.Debug(),
		Transport: newDryRunTransport(),
	}), nil
}

// next advances every bound source by exactly one input item and returns the
// consolidated arguments. io.EOF from any source ends iteration; an
// unresolved parameter yields an item-level cmderrors.ErrNoMatchesFound.
func (r *Runner) next() (Args, error) {
	args := Args{Index: r.index}
	var input any
	for _, s := range r.sources {
		if err := s.advance(&args, &input); err != nil {
			return args, err
		}
	}
	args.Input = input
	r.index++
	return args, nil
}

// producer wraps next() as a job producer with the termination rule:
// piped input and self-ending sources (flag-provided params) stop via io.EOF
// on their own, capped at --maxJobs; a command with neither runs exactly
// once, as its remaining sources would repeat forever.
func (r *Runner) producer(call func(context.Context, Args) output.Seq) Producer {
	produce := Producer(func() (Job, error) {
		args, err := r.next()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return call(ctx, args)
		}, nil
	})

	iterates := r.InputIterators.Total > 0
	for _, s := range r.sources {
		iterates = iterates || s.iterates()
	}
	if iterates {
		return Limit(produce, r.Config.GetMaxJobs())
	}
	return Limit(produce, 1)
}

// Run drains the input, invokes call once per input item (on a worker pool
// when --workers > 1) and renders every yielded document through the
// standard stages: error collection -> --filter -> --outputTemplate ->
// --select -> --output encoding (teed to --outputFile). The returned error
// reflects item-level failures so the exit code is non-zero on partial
// failure.
func (r *Runner) Run(call func(ctx context.Context, args Args) output.Seq) error {
	cfg := r.Config

	ctx := r.Cmd.Context()
	if cfg.ShouldUseDryRun(r.Cmd.CommandPath()) {
		// Dry run reports the calls that would be made without returning the
		// library's mocked responses (which are aimed at tests).
		ctx = apiv2.WithDryRun(ctx, true)
		ctx = ctxhelpers.WithMockResponses(ctx, false)
	}
	ctx = WithProcessingMode(ctx, r.Cmd)

	collector := &ErrorCollector{Max: cfg.AbortOnErrorCount()}
	stages, err := CommonStages(cfg)
	if err != nil {
		return err
	}
	renderer, err := NewRenderer(r.Cmd.OutOrStdout(), cfg)
	if err != nil {
		return err
	}

	err = output.Render(
		ctx,
		Stream(ctx, cfg.GetWorkers(), r.producer(call)),
		renderer,
		append([]output.Stage{collector.Stage()}, stages...)...,
	)
	if err != nil {
		return err
	}
	return collector.Err()
}
