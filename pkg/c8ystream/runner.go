package c8ystream

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/stream"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/filter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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

	userErrorPath     string      // see UserErrorRequests
	directNameLookup  bool        // see DirectNameLookup
	dropPageSize      atomic.Bool // a list sends no page size, see noPageSize
	noDefaultPageSize bool        // see NoDefaultPageSize
	getPageSize       int         // page size of the GET requests of a call (0: none), see defaultPageSizeHook

	viewMu          sync.Mutex           // guards viewResp against the response middleware
	viewResp        *viewResponse        // the latest response, which --view auto matches
	viewDefinition  *dataview.Definition // the view applied to the output; its column widths apply to tables
	consoleRenderer *consoleRenderer     // the table/json renderer; nil for csv/tsv
	index           int64
	fileReceived    atomic.Bool // a response with a body was received (--outputFile)

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
	f.SetRunningCommand(cmd)
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
	var driver iterator.Iterator
	err := iterator.ErrNoPipeInput
	if !r.stdinDisabled() {
		driver, err = iterator.NewJSONPipeIterator(r.Cmd.InOrStdin(), &iterator.PipeOptions{Mode: stream.ModeAuto})
	}
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
	pipeOpts.Disabled = pipeOpts.Disabled || r.stdinDisabled()
	iter, err := flags.NewFlagWithPipeIterator(r.Cmd, pipeOpts, true)
	if err != nil {
		if name == "device" && errors.Is(err, fileguard.ErrDenied) && r.assert == nil && !devicePathCommand(r.Cmd) {
			// A denied file of the source device filter is a user error (exit
			// code 101), unlike a denied file of a path argument (--id) or of an
			// assert command, as in v1.
			return cmderrors.NewUserError(err)
		}
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

// devicePathCommands are the commands whose --device is a path argument rather
// than a filter or body property; v1 reported a denied file there as an
// unexpected error (exit code 100).
var devicePathCommands = []string{" devices services ", " identity ", " remoteaccess "}

func devicePathCommand(cmd *cobra.Command) bool {
	path := cmd.CommandPath() + " "
	for _, c := range devicePathCommands {
		if strings.Contains(path, c) {
			return true
		}
	}
	return false
}

// stdinDisabled reports whether reading piped input from stdin is turned off
// (--nullInput / settings.defaults.nullInput), e.g. because stdin carries a
// proxied stream. The flags' own values still drive the command.
func (r *Runner) stdinDisabled() bool {
	return r.Config != nil && r.Config.DisableStdin()
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
		Disabled:    r.stdinDisabled(),
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

// Client builds the go-c8y v2 client for this command (see
// cmdutil.Factory.ClientV2, which shell completion shares) and installs the
// Runner's middleware: activity log, --outputTemplate capture and the silent
// success message.
func (r *Runner) Client() (*apiv2.Client, error) {
	client, err := r.Factory.ClientV2()
	if err != nil {
		return nil, err
	}
	if err := r.installActivityLog(client); err != nil {
		return nil, err
	}
	r.installOutputTemplateCapture(client)
	r.installOutputFileCapture(client)
	r.installViewCapture(client)
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
	if progress := r.Factory.IOStreams.ProgressIndicator(); progress.IsEnabled() {
		// show a progress bar for the download of a binary
		ctx = apiv2.WithResponseBodyWrapper(ctx, c8ybinary.DownloadProgress(progress))
	}

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

	// v2.59.0 sent the page size on every GET request, also on the ones which
	// are no list (get by id, count, ...)
	if common, err := cfg.GetOutputCommonOptions(r.Cmd); err == nil && !r.noDefaultPageSize {
		r.getPageSize = common.PageSize
	}

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
	stages, err := commonStages(cfg, nil)
	if err != nil {
		return err
	}
	if len(r.filters) > 0 {
		pred, err := filter.Parse(r.filters...)
		if err != nil {
			return err
		}
		stages = append([]output.Stage{filterStage(cfg, pred)}, stages...)
	}
	// --view: when the user didn't select columns, restrict the documents to
	// the detected/named view's columns (no-op for --view off / the non-TTY
	// default), matching the v1 output path.
	cols := newColumnSelector(cfg)
	vs := r.viewStage(cols)
	if len(cfg.GetJSONSelect()) == 0 && cfg.FlattenJSON() && len(stages) > 0 {
		// --flatten (without --select) is the last common stage. As in v1, where
		// the view was the selection which --flatten flattened, the view comes
		// first: it shapes the nested document (json), or resolves the columns
		// which a tabular renderer applies to the nested document.
		flatten := stages[len(stages)-1]
		stages = stages[: len(stages)-1 : len(stages)-1]
		if vs != nil {
			stages = append(stages, vs)
		}
		switch {
		case selectionCSVActive(cfg):
			// csv/tsv: the columns of the renderer are the flat paths of the
			// document already (also an empty object or array), so a flattened
			// document would lose those
		case columnsInRenderer(cfg):
			stages = append(stages, flattenUnlessColumns(flatten, cols))
		default:
			stages = append(stages, flatten)
		}
	} else if vs != nil {
		stages = append(stages, vs)
	}
	renderer, err := newRenderer(r.Cmd.OutOrStdout(), cfg, cols, r.fileReceivedFunc)
	if err != nil {
		return err
	}
	r.consoleRenderer = findConsoleRenderer(renderer)
	r.applyViewDefinition()

	// --progress: count items as they flow to the renderer (stages run on the
	// single consuming goroutine, so the bar needs no locking).
	progress := newProgress(r.Factory.IOStreams.ErrOut, cfg.ShowProgress())
	pipeline := []output.Stage{collector.Stage()}
	if progress != nil {
		pipeline = append(pipeline, progress.stage())
	}
	pipeline = append(pipeline, stages...)
	// v2.59.0 printed "<", ">" and "&" as the characters, not as their \u escapes
	pipeline = append(pipeline, unescapeHTMLStage())

	err = output.Render(ctx, Stream(ctx, cfg.GetWorkers(), r.logJobs(newProducer())), renderer, pipeline...)
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
			// Deleting an id that does not exist is reported as a 404 error
			// (exit code 4) and creating or assigning one that exists as a 409
			// error, as go-c8y-cli always did. The SDK would skip both.
			// Resolver lookups use runCtx and keep their own behavior.
			callCtx := apiv2.WithConflictAsError(apiv2.WithNotFoundAsError(apiv2.WithApprovalScope(ctx)))
			if r.getPageSize > 0 {
				callCtx = apiv2.WithRequestHook(callCtx, defaultPageSizeHook(r.getPageSize))
			}
			if r.dropPageSize.Load() {
				callCtx = apiv2.WithRequestHook(callCtx, dropNoPageSizeHook())
			}
			seq := r.adaptItem(call(callCtx), item)
			if r.assert != nil {
				seq = r.assert.filter(seq)
			}
			return seq
		})
		if r.submission != nil {
			job = r.submission.withDelays(job)
		}
		// --outputFileRaw writes the unshaped response, so it wraps the job
		// before the output template does
		if rawPath := r.Config.GetOutputFileRaw(); rawPath != "" {
			job = r.wrapRawFileJob(rawPath, job)
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
	// as v2.59.0, which logged the line of the input (<nil> without input)
	if log, err := r.Factory.Logger(); err == nil && log != nil && !r.Factory.IsStreaming() {
		if r.driver != nil {
			log.Debugf("Input line: %s", in.input)
		} else {
			log.Debugf("Input line: %s", nil)
		}
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
