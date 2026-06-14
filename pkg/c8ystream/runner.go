package c8ystream

import (
	"context"
	"errors"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/stream"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/authentication"
	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// Call is one service invocation. It is built (with its arguments already
// resolved) on the serial input-draining goroutine and executed later,
// possibly on a worker pool, so it must not read the runner's per-item state.
type Call func(ctx context.Context) output.Seq

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

	cursor      *iterator.InputCursor
	driver      iterator.Iterator // pipe or primary-flag iterator; nil = run once
	primaryName string            // flag the driver resolves (e.g. "id"); "" for a plain pipe
	query       *flags.QueryTemplate
	body        *mapbuilder.MapBuilder
	runCtx      context.Context // set in Run; exposed to Build via Resolver.Context
	submission  *submission     // confirmation + delays; nil when neither applies
	index       int64
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
	if r.InputIterators.PipeOptions != nil {
		r.InputIterators.PipeOptions.Name = ""
		r.InputIterators.PipeOptions.Disabled = true
	}
	driver, err := iterator.NewJSONPipeIterator(r.Cmd.InOrStdin(), &iterator.PipeOptions{Mode: stream.ModeAuto})
	switch {
	case err == nil:
		r.driver = driver
	case errors.Is(err, iterator.ErrNoPipeInput), errors.Is(err, iterator.ErrEmptyPipeInput):
		// no pipe: run once
	default:
		return err
	}
	return nil
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
	if pipeOpts == nil {
		pipeOpts = &flags.PipelineOptions{Name: name, Property: name, Required: true}
	}
	iter, err := flags.NewFlagWithPipeIterator(r.Cmd, pipeOpts, true)
	if err != nil {
		return err
	}
	if iter != nil {
		r.driver = iter
	}
	r.primaryName = name
	return nil
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

// Body configures the request body document built from the given flag options
// (data flag, templates, piped values). Evaluate it per item in Build with
// in.Body().
func (r *Runner) Body(getters ...flags.GetOption) error {
	r.body = mapbuilder.NewInitializedMapBuilder(true)
	if err := flags.WithBody(r.Cmd, r.body, r.InputIterators, getters...); err != nil {
		return cmderrors.NewUserError(err)
	}
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
	client := apiv2.NewClient(apiv2.ClientOptions{
		BaseURL: legacy.BaseURL.String(),
		Auth: authentication.AuthOptions{
			Tenant:   legacy.TenantName,
			Username: legacy.Username,
			Password: legacy.Password,
			Token:    legacy.Token,
		},
		Debug: r.Config.Verbose() || r.Config.Debug(),
	})
	if err := r.installActivityLog(client); err != nil {
		return nil, err
	}
	return client, nil
}

// Run drains the input one item at a time, calls build to assemble each
// service Call (resolution is serial), executes the calls (on a worker pool
// when --workers > 1) and renders every yielded document through the standard
// stages: error collection -> --filter -> --outputTemplate -> --select ->
// --output encoding (teed to --outputFile). The returned error reflects
// item-level failures so the exit code is non-zero on partial failure.
func (r *Runner) Run(build Build) error {
	cfg := r.Config

	ctx := r.Cmd.Context()
	if cfg.ShouldUseDryRun(r.Cmd.CommandPath()) {
		// Report the calls that would be made without the library's mocked
		// responses (which are aimed at tests). The handler renders each
		// prepared request to stdout in the --dryFormat format.
		ctx = apiv2.WithDryRun(ctx, true)
		ctx = ctxhelpers.WithMockResponses(ctx, false)
		handler, err := r.dryRunHandler()
		if err != nil {
			return err
		}
		ctx = ctxhelpers.WithDryRunHandler(ctx, handler)
	}
	ctx = WithProcessingMode(ctx, r.Cmd, cfg)

	// Confirmation + delays live on the submission, shared across items and
	// carried via context to the per-item Submit calls. Built once here so the
	// prompt counter and "all" answers span the whole batch.
	r.submission = r.newSubmission()
	ctx = withSubmission(ctx, r.submission)

	r.runCtx = ctx

	collector := &ErrorCollector{Max: cfg.AbortOnErrorCount()}
	stages, err := CommonStages(cfg)
	if err != nil {
		return err
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

	err = output.Render(ctx, Stream(ctx, cfg.GetWorkers(), r.producer(build)), renderer, pipeline...)
	if progress != nil {
		progress.done()
	}
	if err != nil {
		return err
	}
	// Convert the collected item errors to a status-aware CommandError, then run
	// it through the standard post-command handling so --silentStatusCodes,
	// --silentExit and error printing behave exactly as on the v1 worker path.
	// Skip it entirely on success — CheckPostCommandError treats a nil error as
	// an unexpected one and would print a spurious message.
	finalErr := r.collapseErrors(collector.Errors)
	if finalErr == nil {
		return nil
	}
	return r.Factory.CheckPostCommandError(finalErr)
}

// producer drains one input item, resolves it into a Call via build, and
// caps iteration: a driver (pipe or primary flag) ends via io.EOF, bounded by
// --maxJobs; with no driver the command runs exactly once.
func (r *Runner) producer(build Build) Producer {
	produce := Producer(func() (Job, error) {
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
		r.index++

		call, err := build(in)
		if err != nil {
			return nil, err
		}
		job := Job(func(ctx context.Context) output.Seq {
			return call(ctx)
		})
		if r.submission != nil {
			job = r.submission.withDelays(job)
		}
		return job, nil
	})

	if r.driver != nil {
		return Limit(produce, r.Config.GetMaxJobs())
	}
	return Limit(produce, 1)
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
