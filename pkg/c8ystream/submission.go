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
	// Target shown in the prompt (best-effort, taken from the session so it does not send any request)
	if session, err := r.Factory.ClientSession(); err == nil {
		if session.Host != "" {
			s.targetInfo = fmt.Sprintf("host %s (tenant %s)", session.Host, session.Tenant)
		} else {
			s.targetInfo = fmt.Sprintf("tenant %s", session.Tenant)
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
	return s.cfg != nil && s.cfg.ConfirmationPossible()
}

// hasDelays reports whether any inter-request delay is configured.
func (s *submission) hasDelays() bool {
	return s.delay > 0 || s.delayBefore > 0
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

	if !s.requiresPrompt(req) {
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

// requiresPrompt reports whether req must be confirmed. A command with a
// semantic method (flags.WithSemanticMethod) asks according to that method, not
// to the method of the request: a read semantic method (e.g. datahub query,
// deviceregistration getCredentials, which send a POST) only asks with
// --confirm, as in v2.59.0.
func (s *submission) requiresPrompt(req *http.Request) bool {
	method := s.semanticMethod
	if method == "" && req != nil {
		method = req.Method
	}
	return s.cfg.ShouldConfirm(method)
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

// Submit runs one single-result service call and renders its result.
//
// Deprecated: return the service result from the Call instead
// (`return client.X.Create(ctx, body)`). Confirmation now happens in the SDK
// transport through the Runner's request approver, for every request.
func Submit[T jsondoc.Unwrapper](ctx context.Context, call func(context.Context) op.Result[T]) output.Seq {
	return FromResult(call(ctx))
}

// SubmitUpload is Submit for streaming/multipart uploads.
//
// Deprecated: return the service result from the Call instead. The request
// approver sees the real upload request before its body is streamed.
func SubmitUpload[T jsondoc.Unwrapper](ctx context.Context, method, id string, call func(context.Context) op.Result[T]) output.Seq {
	return FromResult(call(ctx))
}

// SubmitStatus is Submit for calls with no renderable body.
//
// Deprecated: return the service result from the Call instead.
func SubmitStatus[T any](ctx context.Context, call func(context.Context) op.Result[T]) output.Seq {
	return FromStatus(call(ctx))
}

// SubmitValue is Submit for calls whose body is a plain Go value.
//
// Deprecated: return the service result from the Call instead.
func SubmitValue[T any](ctx context.Context, dry bool, call func(context.Context) op.Result[T]) output.Seq {
	return FromJSONValue(call(ctx), dry)
}

// errSeq yields a single error into the stream.
func errSeq(err error) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		yield(jsondoc.Empty(), err)
	}
}

// emptySeq yields nothing (a skipped item produces no output and no error).
var emptySeq output.Seq = func(yield func(jsondoc.JSONDoc, error) bool) {}
