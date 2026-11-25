package c8ystream

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/stream"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/filter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log/slog"
)

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// decompressTransport gunzips gzip-encoded response bodies (removing the
// Content-Encoding header) so layers above it — most importantly the response
// cache, whose entries are shared with the v1-path commands — always see the
// decoded body, matching what Go's transport produces when no explicit
// Accept-Encoding header suppresses its transparent gzip handling.
type decompressTransport struct {
	base http.RoundTripper
}

func (t decompressTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, gzErr := gzip.NewReader(resp.Body)
		if gzErr != nil {
			// Leave the response untouched (e.g. an empty body); the caller's
			// own decoding will surface any real corruption.
			return resp, nil
		}
		resp.Body = &gunzipReadCloser{gz: gz, raw: resp.Body}
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
		resp.Uncompressed = true
	}
	return resp, nil
}

// gunzipReadCloser reads decompressed content and closes both the gzip reader
// and the underlying network body.
type gunzipReadCloser struct {
	gz  *gzip.Reader
	raw io.ReadCloser
}

func (g *gunzipReadCloser) Read(p []byte) (int, error) { return g.gz.Read(p) }

func (g *gunzipReadCloser) Close() error {
	err := g.gz.Close()
	if rawErr := g.raw.Close(); err == nil {
		err = rawErr
	}
	return err
}

// Call is one service invocation. It is built (with its arguments already
// resolved) on the serial input-draining goroutine and executed later,
// possibly on a worker pool, so it must not read the runner's per-item state.
//
// A Call returns whatever the service method returned: an op.Result, a
// pagination iterator (any op.DocumentSource), or an output.Seq for local
// documents. The Runner turns it into documents in one place (see adapt), so a
// command never chooses how to confirm, page or render its result:
//
//	return func(ctx context.Context) any { return client.Events.Delete(ctx, id) }, nil
type Call func(ctx context.Context) any

// Build maps the resolved values of one input item to the service call that
// uses them. Resolution (the Resolver methods) happens while Build runs — on
// the serial side — and the returned Call captures the resolved values, so a
// command never thinks about query vs body vs path: it just fills the typed
// options/arguments of its service function and returns the call.
type Build func(in *Resolver) (Call, error)

// Runner owns the command-agnostic lifecycle of a v2-based command: the input
// driver and shared cursor, optional query/body builders, context
// construction (dry-run, processing mode), the worker pool, error collection,
// the standard output stages and rendering.
type Runner struct {
	Cmd            *cobra.Command
	Factory        *cmdutil.Factory
	Config         *config.Config
	InputIterators *flags.RequestInputIterators

	cursor       *iterator.InputCursor
	driver       iterator.Iterator // pipe or primary-flag iterator; nil = run once
	primaryName  string            // flag the driver resolves (e.g. "id"); "" for a plain pipe
	pipeOverride bool              // another flag claimed the pipe via `-`/`-.path`
	query        *flags.QueryTemplate
	body         *mapbuilder.MapBuilder
	runCtx       context.Context // set in Run; exposed to Build via Resolver.Context
	submission   *submission     // confirmation + delays; nil when neither applies
	assert       *assertion      // set by Assert; nil for other commands
	filters      []string        // filters of the command itself (see Filter)
	index        int64

	// tmplInput is the current item's raw input, kept pointed at the item being
	// produced so the --outputTemplate evaluation can bind it as input.value.
	// Correct for the serial pipeline (the default); request/response stay empty
	// until a command threads its response metadata.
	tmpl         *mapbuilder.OutputTemplate // parsed --outputTemplate; nil when unset
	tmplFlags    map[string]string
	tmplInput    []byte
	tmplMu       sync.Mutex // guards tmplRequest/tmplResponse against the response middleware
	tmplRequest  map[string]any
	tmplResponse map[string]any
}

// NewRunner resolves the command configuration and pipeline annotations and
// installs the shared input cursor, so any string flag valued `-`/`-.path`
// resolves from the current input item.
func NewRunner(cmd *cobra.Command, f *cmdutil.Factory) (*Runner, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	flags.WithOptions(cmd, flags.WithRuntimePipelineProperty())
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return nil, err
	}
	r := &Runner{
		Cmd:            cmd,
		Factory:        f,
		Config:         cfg,
		InputIterators: inputIterators,
		cursor:         &iterator.InputCursor{},
	}
	inputIterators.InputCursor = r.cursor
	return r, nil
}

// Input reads piped stdin one item per line (the default driver). With no
// pipe the command runs once. Other flags can reference the current item via
// `-`/`-.path`.
func (r *Runner) Input() error {
	// This driver owns stdin, so neutralise the legacy single-flag pipe
	// binding (e.g. WithExtendedPipelineSupport's target) to avoid a second
	// competing reader. PipeOptions stays non-nil — several getters dereference
	// it under only an inputIterators nil-check — but with no name nothing
	// binds to it.
	pipeName := ""
	if r.InputIterators.PipeOptions != nil {
		pipeName = r.InputIterators.PipeOptions.Name
		r.InputIterators.PipeOptions.Name = ""
		r.InputIterators.PipeOptions.Disabled = true
	}
	driver, err := iterator.NewJSONPipeIterator(r.Cmd.InOrStdin(), &iterator.PipeOptions{Mode: stream.ModeAuto})
	switch {
	case err == nil:
		r.driver = driver
	case errors.Is(err, iterator.ErrNoPipeInput), errors.Is(err, iterator.ErrEmptyPipeInput):
		// No pipe, so stdin cannot be double-read: restore the per-flag pipe
		// binding (Disabled stays true, so getters matching PipeOptions.Name
		// read the flag's own values — including a file reference, one item
		// per line — instead of stdin, matching v1). The flag's values then
		// drive iteration when they are multi-valued or a file; otherwise the
		// command runs once.
		if r.InputIterators.PipeOptions != nil {
			r.InputIterators.PipeOptions.Name = pipeName
		}
		r.installFlagValuesDriver()
	default:
		return err
	}
	return nil
}

// installFlagValuesDriver makes the pipe-annotated flag's own values the input
// driver when there is no pipe and the flag holds more than one literal value
// (a slice) or a file reference (one item per line). This restores v1's flag
// fan-out for generic-pipe (Runner.Input) commands: the per-flag driver
// (InputFlag) already iterates flag values via NewFlagWithPipeIterator, but
// Input-driven commands would otherwise run once and silently use only the
// first value. Single literal values keep the plain run-once path so behaviour
// (and template input.value visibility) is unchanged; the body picks the
// per-item value up via the restored per-flag pipe binding (getters) or
// Resolver.ApplyOverrideSlice, which cycles the flag values by item index.
func (r *Runner) installFlagValuesDriver() {
	opts, err := flags.GetPipeOptionsFromAnnotation(r.Cmd)
	if err != nil || opts == nil || opts.Name == "" {
		return
	}
	f := r.Cmd.Flags().Lookup(opts.Name)
	if f == nil || !r.Cmd.Flags().Changed(opts.Name) {
		return
	}
	values, err := flags.GetFlagStringValues(r.Cmd, opts.Name)
	if err != nil {
		return
	}
	literals := make([]string, 0, len(values))
	for _, v := range values {
		if _, isRef := iterator.InputReference(v); v == "" || isRef {
			continue
		}
		literals = append(literals, v)
	}
	if len(literals) == 1 && isInputFile(literals[0]) {
		if iter, err := iterator.NewFileContentsIterator(literals[0]); err == nil {
			r.driver = iter
		}
		return
	}
	if len(literals) > 1 && strings.Contains(f.Value.Type(), "Slice") {
		r.driver = iterator.NewSliceIterator(literals)
	}
}

// isInputFile reports whether the value references an existing file (a path,
// not a bare name), mirroring the flags package's input-file detection.
func isInputFile(p string) bool {
	if !strings.ContainsAny(p, `/\`) {
		return false
	}
	fp, err := os.Stat(p)
	return err == nil && !fp.IsDir()
}

// InputFlag makes a flag the iterating driver: piped items feed it (extracting
// the flag's pipeline property/aliases from JSON), or its own slice/file values
// drive iteration when there is no pipe (e.g. `--id 1,2,3`). The value is read
// raw — no name resolution; pass it to a service method (or NameOrID) and let
// go-c8y resolve. Read it back in Build with in.String(name).
func (r *Runner) InputFlag(name string) error {
	pipeOpts, err := flags.GetPipeOptionsFromAnnotation(r.Cmd)
	if err != nil {
		return err
	}
	if pipeOpts == nil || pipeOpts.Name == "" {
		// No pipeline annotation: the flag itself drives the loop and piped
		// items are read by the flag's own name.
		pipeOpts = &flags.PipelineOptions{Name: name, Property: name}
	}
	iter, err := flags.NewFlagWithPipeIterator(r.Cmd, pipeOpts, true)
	if err != nil {
		return err
	}
	if iter != nil {
		// Mirror v1 (flags.WithStringSliceValues): an explicit --<name> value
		// fills in whenever a piped item yields no value for the bound property,
		// so e.g. `... | c8y devices update --id 0` still targets id 0 when the
		// piped object carries no id/deviceId/source.id. The bound pipe keeps
		// driving the loop length; the override only supplies the value, so an id
		// present in the item still wins.
		if values, verr := flags.GetFlagStringValues(r.Cmd, name); verr == nil && r.Cmd.Flags().Changed(name) && len(values) > 0 && iter.IsBound() && !anyInputReference(values) {
			iter = iterator.NewOverrideIterator(iter, iterator.NewInfiniteSliceIterator(values))
		}
		r.driver = iter
	}
	r.primaryName = name
	r.detectPipeOverride()
	return nil
}

// anyInputReference reports whether any value is a `-`/`-.path` pipe reference
// rather than a literal override value (the latter wins over the pipe; the
// former IS the pipe and must be resolved by the driver, not overridden).
func anyInputReference(values []string) bool {
	for _, v := range values {
		if _, ok := iterator.InputReference(v); ok {
			return true
		}
	}
	return false
}

// detectPipeOverride records whether the user mapped the pipeline onto a flag
// other than the driver's, e.g. `c8y applications list --user -` (the driver is
// "type"). When set, the driver still reads stdin to advance iteration and the
// cursor, but its own flag value is left unset so only the explicitly mapped
// flag(s) take the piped value — matching v1, where `--<flag> -` overrides the
// command's default pipeline target.
func (r *Runner) detectPipeOverride() {
	r.Cmd.Flags().Visit(func(f *pflag.Flag) {
		if r.pipeOverride || f.Name == r.primaryName {
			return
		}
		vals, err := flags.GetFlagStringValues(r.Cmd, f.Name)
		if err != nil {
			return
		}
		for _, v := range vals {
			if _, ok := iterator.InputReference(v); ok {
				r.pipeOverride = true
				return
			}
		}
	})
}

// InputRaw makes a flag the iterating driver like InputFlag, but yields each
// item whole (no pipeline-property extraction) and keeps only JSON lines.
// Use for passthrough commands (e.g. util show) where the item itself is the
// payload. Read it back in Build with in.Input().
func (r *Runner) InputRaw(name string) error {
	iter, err := flags.NewFlagWithPipeIterator(r.Cmd, &flags.PipelineOptions{
		Name:        name,
		InputFilter: flags.FilterJsonLines,
		Required:    true,
	}, true)
	if err != nil {
		return err
	}
	if iter != nil {
		r.driver = iter
	}
	r.primaryName = name
	return nil
}

// InputPipeline makes a flag the iterating driver with the given pipeline
// options, for commands that read their input as text rather than as JSON
// items (a line filter, a formatter or a text mode). Piped lines feed the
// flag; without a pipe the flag's own values (or a file it names) do. With no
// input and Required unset, the command runs once with an empty input. Read
// the item back in Build with in.Input().
func (r *Runner) InputPipeline(opts flags.PipelineOptions) error {
	if r.InputIterators.PipeOptions != nil && r.InputIterators.PipeOptions.Disabled {
		opts.Disabled = true
	}
	iter, err := flags.NewFlagWithPipeIterator(r.Cmd, &opts, true)
	if err != nil {
		return err
	}
	if iter != nil {
		r.driver = iter
	}
	r.primaryName = opts.Name
	return nil
}

// Query configures the Cumulocity query expression assembled from the given
// flag options (e.g. flags.WithCumulocityQuery). Evaluate it per item in Build
// with in.Query().
func (r *Runner) Query(getters ...flags.GetOption) error {
	r.query = flags.NewQueryTemplate()
	if err := flags.WithQueryParameters(r.Cmd, r.query, r.InputIterators, getters...); err != nil {
		return cmderrors.NewUserError(err)
	}
	return nil
}

// Filter adds client side filters which the command itself applies to its
// output (e.g. the filter of an extension command), in addition to the
// user's --filter values. Each expression uses the --filter syntax, and a
// document must match all of them.
func (r *Runner) Filter(expressions ...string) error {
	if len(expressions) == 0 {
		return nil
	}
	if _, err := filter.Parse(expressions...); err != nil {
		return cmderrors.NewUserError(err)
	}
	r.filters = append(r.filters, expressions...)
	return nil
}

// Body configures the request body document built from the given flag options
// (data flag, templates, piped values). Evaluate it per item in Build with
// in.Body().
func (r *Runner) Body(getters ...flags.GetOption) error {
	r.body = mapbuilder.NewInitializedMapBuilder(true)
	if err := flags.WithBody(r.Cmd, r.body, r.InputIterators, getters...); err != nil {
		return cmderrors.NewUserError(err)
	}
	r.bindPipeBodyProperty()
	return nil
}

// Client builds a go-c8y v2 client from the session via the legacy factory
// client, which owns credential resolution (encrypted session decryption,
// passphrase prompting, token refresh), so both clients always agree.
// --verbose/--debug enable the v2 client's request/response logging.
func (r *Runner) Client() (*apiv2.Client, error) {
	legacy, err := r.Factory.Client()
	if err != nil {
		return nil, err
	}
	opts := apiv2.ClientOptions{
		BaseURL: legacy.BaseURL.String(),
		Auth: authentication.AuthOptions{
			Tenant:   legacy.TenantName,
			Username: legacy.Username,
			Password: legacy.Password,
			Token:    legacy.Token,
		},
		Debug:   r.Config.Verbose() || r.Config.Debug(),
		Timeout: r.Config.RequestTimeout(),
	}

	// Mirror the legacy client's negotiated authorization method. Without this
	// a password-only session makes the v2 client exchange the password for a
	// bearer token (POST /tenant/oauth/token) on EVERY command, where v1 sent
	// Authorization: Basic directly — slower, chattier, and a poisoning hazard
	// when the exchange is replayed (revoked tokens). Bearer sessions keep the
	// token-source behaviour so expired tokens still refresh.
	switch legacy.AuthorizationType {
	case c8y.AuthTypeBasic:
		opts.Auth.AuthType = []authentication.AuthType{authentication.AuthTypeBasic}
	case c8y.AuthTypeBearer:
		opts.Auth.AuthType = []authentication.AuthType{authentication.AuthTypeBearer}
	}

	// Install the HTTP response cache exactly as the legacy factory client does
	// (pkg/cmd/factory/c8yclient.go). The cache is a RoundTripper wrapper, so it
	// becomes the v2 client's base transport; apiv2.NewClient further wraps it in
	// DryRunTransport, giving DryRun -> cache -> network ordering (a dry run
	// short-circuits before ever touching the cache). Without this the v2 path
	// silently skipped caching entirely.
	if r.Config.CacheEnabled() && r.Config.CacheTTL() > 0 {
		cachableMethods := r.Config.CacheMethods()
		// The cache must store DECODED bodies. The v2 resty client sends an
		// explicit Accept-Encoding header, which disables Go's transparent gzip
		// handling, so responses reach this transport still compressed with
		// Content-Encoding: gzip. Cached as-is they poison the shared cache dir:
		// the v1-path commands (whose transport decompresses transparently before
		// their cache wrapper) replay the entry without decompressing and parse
		// gzip bytes as JSON. decompressTransport normalises the response before
		// the cache sees it.
		opts.Transport = c8y.CacheResponse(
			r.Config.CacheTTL(),
			r.Config.CacheDir(),
			func(req *http.Request) bool {
				// Never cache authentication requests. The v2 client acquires
				// bearer tokens through this same transport (v1 never did), and
				// a cached token response outlives revocation: the token-source
				// refresh POST is then also served from the cache, so the client
				// re-fetches the same revoked token and every request 401s until
				// the cache TTL expires.
				if strings.Contains(req.URL.Path, "/tenant/oauth") {
					return false
				}
				return strings.Contains(cachableMethods, req.Method)
			},
			c8y.CacheOptions{
				ExcludeAuth: !r.Config.CacheKeyIncludeAuth(),
				ExcludeHost: !r.Config.CacheKeyIncludeHost(),
				Mode:        r.Config.CacheMode(),
				BodyKeys:    r.Config.CacheBodyKeys(),
			},
		)(decompressTransport{base: http.DefaultTransport})
	}

	// v1's SDK logged "Sending request: <method> <url>" (session values masked)
	// before every send. Wrap the transport outside the cache so the line also
	// appears for cache hits, exactly as v1's pre-transport log did; dry runs
	// still skip it because DryRunTransport short-circuits before this layer.
	if logg, logErr := r.Factory.Logger(); logErr == nil && logg != nil {
		base := opts.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		opts.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			logg.Infof("Sending request: %s %s", req.Method, r.Config.HideSensitiveInformationIfActive(legacy, req.URL.String()))
			return base.RoundTrip(req)
		})
	}

	// The SDK's HTTP layer logs its warnings (e.g. credentials over plain
	// HTTP) at info level and the realtime client logs every websocket frame
	// at info level: show them only with --verbose/--debug.
	level := slog.LevelWarn
	if r.Config.Verbose() || r.Config.Debug() {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(r.Factory.IOStreams.ErrOut, &slog.HandlerOptions{Level: level}))
	opts.Logger = logger

	client := apiv2.NewClient(opts)
	client.Realtime().SetLogger(logger.With("component", "realtime"))
	if err := r.installActivityLog(client); err != nil {
		return nil, err
	}
	r.installOutputTemplateCapture(client)
	r.installSilentSuccessMessage(client)
	return client, nil
}

// Run drains the input one item at a time, calls build to assemble each
// service Call (resolution is serial), executes the calls (on a worker pool
// when --workers > 1) and renders every yielded document through the standard
// stages: error collection -> --filter -> --outputTemplate -> --select ->
// --output encoding (teed to --outputFile). The returned error reflects
// item-level failures so the exit code is non-zero on partial failure.
func (r *Runner) Run(build Build) error {
	return r.run(func() Producer { return r.producer(build) })
}

// run runs the jobs of the producer which newProducer returns (once the run
// context is set) through the standard pipeline.
func (r *Runner) run(newProducer func() Producer) error {
	cfg := r.Config

	ctx := r.Cmd.Context()
	if cfg.ShouldUseDryRun(r.Cmd.CommandPath()) {
		// Report the calls that would be made without the library's mocked
		// responses (which are aimed at tests). The handler renders each
		// prepared request to stdout in the --dryFormat format.
		ctx = apiv2.WithDryRun(ctx, true)
		ctx = apiv2.WithMockResponses(ctx, false)
		handler, err := r.dryRunHandler()
		if err != nil {
			return err
		}
		ctx = apiv2.WithDryRunHandler(ctx, handler)
	}
	ctx = WithProcessingMode(ctx, r.Cmd, cfg)
	ctx = WithExtraRequestOptions(ctx, cfg)

	// Confirmation + delays live on the submission, shared across items and
	// carried via context to the per-item Submit calls. Built once here so the
	// prompt counter and "all" answers span the whole batch.
	r.submission = r.newSubmission()
	ctx = withSubmission(ctx, r.submission)
	if r.submission != nil && r.submission.confirmActive() {
		// Confirmation is a property of the request, not of the command: the SDK
		// asks the approver for every request that changes data, after names
		// are resolved and right before it is sent. No command can skip it.
		ctx = apiv2.WithRequestApprover(ctx, func(req *http.Request) (bool, error) {
			ok, err := r.submission.confirm(req)
			if err != nil {
				// Mark the prompt's own error so adapt can report it as is,
				// without the transport/HTTP wrapping it picks up on the way.
				return false, approvalError{err}
			}
			return ok, nil
		})
	}

	r.runCtx = ctx

	collector := &ErrorCollector{Max: cfg.AbortOnErrorCount()}
	if cfg.WithError() && (cfg.GetOutputTemplate() != "" || len(cfg.GetJSONSelect()) > 0) {
		// --withError with shaping: route each item's error body to stdout as a
		// document so --outputTemplate/--select can process it, matching v1.
		// Plain --withError skips the pipeline (whose renderer sorts JSON keys)
		// and lets the post-command handler write the error document verbatim
		// (CommandError.JSONString field order), byte-identical to v1.
		collector.Emit = r.errorDocument
	}
	// Each item error is reported as it occurs through the standard post-command
	// handling (silent-status-code/silent-exit aware, Processed-marked so the
	// final pass never reprints), exactly like v1's per-job checkError. The
	// processed results feed the single-error collapse below.
	var processedErrs []error
	collector.Log = func(err error) {
		converted := r.toServerError(err)
		if ce, ok := converted.(cmderrors.CommandError); ok && collector.Emit != nil {
			// The pipeline already emitted the error document (--withError).
			ce.OutputEmitted = true
			converted = ce
		}
		processedErrs = append(processedErrs, r.Factory.CheckPostCommandError(converted))
	}
	// --outputTemplate is applied per job by the producer (wrapTemplateJob), not
	// as a per-document stage: v1 evaluated the template once per request.
	tmpl, err := r.prepareOutputTemplate()
	if err != nil {
		return err
	}
	r.tmpl = tmpl
	if r.tmpl != nil {
		r.tmplFlags = r.commandFlags()
	}
	stages, err := CommonStages(cfg, nil)
	if err != nil {
		return err
	}
	if len(r.filters) > 0 {
		pred, err := filter.Parse(r.filters...)
		if err != nil {
			return err
		}
		stages = append([]output.Stage{output.Filter(pred)}, stages...)
	}
	// --view: when the user didn't select columns, restrict the documents to
	// the detected/named view's columns (no-op for --view off / the non-TTY
	// default), matching the v1 output path.
	if vs := r.viewStage(); vs != nil {
		stages = append(stages, vs)
	}
	renderer, err := NewRenderer(r.Cmd.OutOrStdout(), cfg)
	if err != nil {
		return err
	}

	// --progress: count items as they flow to the renderer (stages run on the
	// single consuming goroutine, so the bar needs no locking).
	progress := newProgress(r.Factory.IOStreams.ErrOut, cfg.ShowProgress())
	pipeline := []output.Stage{collector.Stage()}
	// --outputFileRaw tees the unshaped document to disk, so it runs before the
	// filter/template/select stages (it sees what the server returned, not the
	// rendered shape), matching v1.
	if rawPath := cfg.GetOutputFileRaw(); rawPath != "" {
		pipeline = append(pipeline, rawFileStage(rawPath))
	}
	if progress != nil {
		pipeline = append(pipeline, progress.stage())
	}
	pipeline = append(pipeline, stages...)

	err = output.Render(ctx, Stream(ctx, cfg.GetWorkers(), newProducer()), renderer, pipeline...)
	if progress != nil {
		progress.done()
	}
	if err != nil {
		// A stream-level abort (e.g. --abortOnErrors reached, exit 103) — the item
		// errors themselves were already reported by collector.Log; run the abort
		// through the post-command handling for its message and exit code.
		return r.finish(err)
	}
	// Reduce the collected item errors to the command result. A single failure
	// resolves to its already-processed error (printed by collector.Log, exit
	// code preserved, nil under --silentExit); multiple failures resolve to the
	// aggregate — matching the v1 worker path.
	var finalErr error
	if len(collector.Errors) == 1 && len(processedErrs) == 1 {
		finalErr = processedErrs[0]
	} else {
		finalErr = r.collapseErrors(collector.Errors)
	}
	if finalErr == nil {
		return nil
	}
	if collector.Emit != nil {
		// The error body was already emitted to stdout as a document by the
		// pipeline (--withError); tell the post-command handler not to write it
		// again so it only sets the exit code and logs to stderr.
		if ce, ok := finalErr.(cmderrors.CommandError); ok {
			ce.OutputEmitted = true
			finalErr = ce
		}
	}
	return r.finish(finalErr)
}

// finish runs the command's final error through the post-command handling.
func (r *Runner) finish(err error) error {
	err = r.Factory.CheckPostCommandError(err)
	if r.assert != nil {
		return r.assert.exitCode(err)
	}
	return err
}

// producer drains one input item, resolves it into a Call via build, and
// caps iteration: a driver (pipe or primary flag) ends via io.EOF, bounded by
// --maxJobs; with no driver the command runs exactly once.
func (r *Runner) producer(build Build) Producer {
	produce := Producer(func() (Job, error) {
		in, err := r.nextItem()
		if err != nil {
			return nil, err
		}

		call, err := build(in)
		if err != nil {
			// A per-item build/validation error (e.g. a missing required body
			// property) is item-level: surface it as a job that yields the error
			// so the collector records it and the batch continues with the next
			// item, matching the v1 worker semantics. Returning (nil, err) here
			// would instead abort the whole stream after the first bad item.
			itemErr := err
			return Job(func(context.Context) output.Seq {
				return func(yield func(jsondoc.JSONDoc, error) bool) {
					yield(jsondoc.Empty(), itemErr)
				}
			}), nil
		}
		item := in.primary
		job := Job(func(ctx context.Context) output.Seq {
			// One approval scope per input item: an item whose call sends
			// several requests (upload then update, ...) is confirmed once.
			seq := r.adaptItem(call(apiv2.WithApprovalScope(ctx)), item)
			if r.assert != nil {
				seq = r.assert.filter(seq)
			}
			return seq
		})
		if r.submission != nil {
			job = r.submission.withDelays(job)
		}
		if r.tmpl != nil {
			job = r.wrapTemplateJob(r.tmpl, r.tmplFlags, job)
		}
		return job, nil
	})

	if r.driver != nil {
		return Limit(produce, r.Config.GetMaxJobs())
	}
	return Limit(produce, 1)
}

// nextItem reads the next input item from the driver (io.EOF when there is
// none left) and returns its resolver. Without a driver it returns an item
// with no input; the caller decides how often to run.
func (r *Runner) nextItem() (*Resolver, error) {
	in := &Resolver{r: r, ctx: r.runCtx, index: r.index}
	if r.driver != nil {
		value, raw, err := r.driver.GetNext()
		if err != nil {
			return nil, err
		}
		in.input = toBytes(raw)
		in.primary = string(value)
		r.cursor.Set(in.input)
	}
	// Point the output-template input at this item. In the serial pipeline
	// (workers<=1) the item's documents are rendered before the next item is
	// produced, so the stage reads the matching input.
	r.tmplInput = in.input
	r.index++
	return in, nil
}

func toBytes(v any) []byte {
	switch b := v.(type) {
	case []byte:
		return b
	case string:
		return []byte(b)
	case nil:
		return nil
	default:
		return fmt.Appendf(nil, "%v", v)
	}
}
