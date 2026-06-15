package c8ystream

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	ctxhelpers "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/contexthelpers"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// submission carries the cross-cutting concerns that sit around a service call
// on the producer/submission side: serialized confirmation prompts and the
// inter-request delays. It is built once per command from the session
// configuration and shared (so "yes to all" / "no to all" and the prompt
// counter span every item), then threaded to the per-item Submit calls via the
// context. The post-execution concerns (activity log, --outputFileRaw, silent
// status codes, views) live elsewhere — at the transport, the renderer and the
// error collector — because they need the response, not the prepared request.
type submission struct {
	cfg            *config.Config
	semanticMethod string
	confirmText    string
	targetInfo     string // "host <fqdn> (tenant <id>)" shown in the prompt
	delay          time.Duration
	delayBefore    time.Duration

	mu          sync.Mutex
	jobID       int64
	skipConfirm bool // user answered "yes to all"
	aborted     bool // user answered "no to all": skip every remaining item
}

// newSubmission builds the submission for a command from the session
// configuration. It returns nil when nothing it governs is active (no
// confirmation, no delays) so the hot path stays allocation- and check-free.
func (r *Runner) newSubmission() *submission {
	cfg := r.Config
	s := &submission{
		cfg:            cfg,
		semanticMethod: flags.GetSemanticMethodFromAnnotation(r.Cmd),
		confirmText:    cfg.ConfirmText(),
		delay:          cfg.WorkerDelay(),
		delayBefore:    cfg.WorkerDelayBefore(),
	}
	if !s.confirmActive() && !s.hasDelays() {
		return nil
	}
	// Target shown in the prompt — best-effort; the client is already built by
	// the time Run executes, so this does not trigger a fresh login.
	if client, err := r.Factory.Client(); err == nil && client != nil {
		if client.BaseURL != nil {
			s.targetInfo = fmt.Sprintf("host %s (tenant %s)", client.GetHostname(), client.TenantName)
		} else {
			s.targetInfo = fmt.Sprintf("tenant %s", client.TenantName)
		}
	}
	return s
}

type submissionKey struct{}

func withSubmission(ctx context.Context, s *submission) context.Context {
	if s == nil {
		return ctx
	}
	return context.WithValue(ctx, submissionKey{}, s)
}

func submissionFrom(ctx context.Context) *submission {
	s, _ := ctx.Value(submissionKey{}).(*submission)
	return s
}

// confirmActive reports whether confirmation prompting could apply at all
// (interactive, not CI/--force/--dry). When false the Submit fast path skips
// the deferred-execution prepare entirely.
func (s *submission) confirmActive() bool {
	return s.cfg != nil && s.cfg.ShouldConfirm()
}

// hasDelays reports whether any inter-request delay is configured.
func (s *submission) hasDelays() bool {
	return s.delay > 0 || s.delayBefore > 0
}

// prepareContext returns a context that makes the next service call build and
// return its prepared *http.Request without sending it (deferred execution).
// A no-op dry-run handler is attached so the transport's dry-run path stays
// silent during this prepare-only round (the real --dry handler, when set,
// already lives on the run context and confirmation is disabled under --dry).
func (s *submission) prepareContext(ctx context.Context) context.Context {
	ctx = ctxhelpers.WithDeferredExecution(ctx, true)
	ctx = ctxhelpers.WithDryRunHandler(ctx, func(*http.Request) {})
	return ctx
}

// confirm prompts for the prepared request and reports whether to proceed. It
// is serialized so prompts never interleave and the "all" answers are shared
// across every item, reproducing the v1 worker semantics.
func (s *submission) confirm(req *http.Request) (proceed bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.aborted {
		return false, nil // a previous "no to all" cancelled the rest
	}
	if s.skipConfirm {
		return true, nil
	}

	method := s.semanticMethod
	if method == "" && req != nil {
		method = req.Method
	}
	if !s.cfg.ShouldConfirm(method) {
		return true, nil // this method is not in the confirmation set
	}

	s.jobID++
	message := s.confirmMessage(s.operationText(), req)
	result, promptErr := prompt.Confirm(fmt.Sprintf("(job: %d)", s.jobID), message, s.targetInfo, prompt.ConfirmYes.String(), false)
	switch result {
	case prompt.ConfirmYesToAll:
		s.skipConfirm = true
		return true, nil
	case prompt.ConfirmYes:
		return true, nil
	case prompt.ConfirmNoToAll:
		// Cancel this and every remaining item; surface the abort once.
		s.aborted = true
		return false, promptErr
	default: // ConfirmNo: skip just this item, recorded as an item error
		return false, promptErr
	}
}

// operationText builds the verb shown in the prompt: the configured
// confirmText, or the command derived from the invocation (e.g. "delete
// device"), matching v1.
func (s *submission) operationText() string {
	if s.confirmText != "" {
		return s.confirmText
	}
	if len(os.Args) > 2 {
		return fmt.Sprintf("%s %s", os.Args[2], strings.TrimRight(os.Args[1], "s"))
	}
	return "Execute command"
}

// confirmMessage appends the request target ("[id=...]") to the operation
// text. The id is taken from the request path (the resolved, about-to-be-sent
// request), so a name → id lookup is already reflected.
func (s *submission) confirmMessage(prefix string, req *http.Request) string {
	id := ""
	if req != nil {
		for part := range strings.SplitSeq(req.URL.Path, "/") {
			if part != "" && c8ydata.IsID(part) {
				id = part
			}
		}
	}
	if id != "" {
		return fmt.Sprintf("%s [id=%s]", prefix, id)
	}
	return prefix
}

// withDelays wraps a job so it sleeps --delayBefore before running and --delay
// after, reproducing the v1 worker timing. The sleeps honour context
// cancellation so a stopped consumer is not held up.
func (s *submission) withDelays(job Job) Job {
	if !s.hasDelays() {
		return job
	}
	return func(ctx context.Context) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			if !sleep(ctx, s.delayBefore) {
				return
			}
			for doc, err := range job(ctx) {
				if !yield(doc, err) {
					return
				}
			}
			sleep(ctx, s.delay)
		}
	}
}

// sleep waits for d, returning false if the context is cancelled first.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// Submit runs one single-result service call with the confirmation lifecycle.
// When confirmation is inactive it is the plain FromResult fast path. When
// active it prepares the request (deferred execution, no send), prompts
// serially, and only then executes for real — so the mutating request is never
// sent before the user agrees.
func Submit[T jsondoc.Unwrapper](ctx context.Context, call func(context.Context) op.Result[T]) output.Seq {
	s := submissionFrom(ctx)
	if s == nil || !s.confirmActive() {
		return FromResult(call(ctx))
	}
	prep := call(s.prepareContext(ctx))
	proceed, err := s.confirm(prep.Request)
	if err != nil {
		return errSeq(err)
	}
	if !proceed {
		return emptySeq
	}
	if prep.IsDeferred() {
		return FromResult(prep.Execute(ctx))
	}
	return FromResult(prep)
}

// SubmitUpload is Submit for streaming/multipart uploads. Submit prepares the
// request via deferred execution to show it in the confirmation prompt, but a
// multipart/streaming body is encoded through an io.Pipe that has no reader
// during the prepare-only round (its dry-run handler is a no-op), so the writer
// blocks and the command hangs before the prompt ever appears. SubmitUpload
// instead confirms from a lightweight request built from the known method and
// resource id (no body), then runs the real upload only if the user agrees.
// Pass id="" for create (POST to a collection) and the resource id for update.
func SubmitUpload[T jsondoc.Unwrapper](ctx context.Context, method, id string, call func(context.Context) op.Result[T]) output.Seq {
	s := submissionFrom(ctx)
	if s == nil || !s.confirmActive() {
		return FromResult(call(ctx))
	}
	// A minimal request carries just the method (drives ShouldConfirm) and the
	// id in the path (shown in the prompt) — never the body.
	req, _ := http.NewRequest(method, "/"+id, nil)
	proceed, err := s.confirm(req)
	if err != nil {
		return errSeq(err)
	}
	if !proceed {
		return emptySeq
	}
	return FromResult(call(ctx))
}

// SubmitStatus is Submit for calls with no renderable body (e.g. Delete
// returning NoContent): same confirmation lifecycle, FromStatus rendering.
func SubmitStatus[T any](ctx context.Context, call func(context.Context) op.Result[T]) output.Seq {
	s := submissionFrom(ctx)
	if s == nil || !s.confirmActive() {
		return FromStatus(call(ctx))
	}
	prep := call(s.prepareContext(ctx))
	proceed, err := s.confirm(prep.Request)
	if err != nil {
		return errSeq(err)
	}
	if !proceed {
		return emptySeq
	}
	if prep.IsDeferred() {
		return FromStatus(prep.Execute(ctx))
	}
	return FromStatus(prep)
}

// errSeq yields a single error into the stream.
func errSeq(err error) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		yield(jsondoc.Empty(), err)
	}
}

// emptySeq yields nothing (a skipped item produces no output and no error).
var emptySeq output.Seq = func(yield func(jsondoc.JSONDoc, error) bool) {}
