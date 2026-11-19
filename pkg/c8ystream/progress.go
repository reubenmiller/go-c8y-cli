package c8ystream

import (
	"io"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

// progress is a minimal --progress indicator: a single spinner on stderr that
// counts items as they are rendered. It uses mpb directly (the pkg/progressbar
// wrapper is worker-pool shaped and its Wait is self-recursive).
type progress struct {
	p   *mpb.Progress
	bar *mpb.Bar
}

// newProgress returns a progress indicator, or nil when disabled (so callers
// branch on nil rather than carry a no-op).
func newProgress(w io.Writer, enabled bool) *progress {
	if !enabled || w == nil {
		return nil
	}
	p := mpb.New(mpb.WithOutput(w))
	// Indeterminate total (-1) so Increment never auto-completes the spinner
	// before done() collapses it; otherwise the count decorator never renders.
	bar := p.AddSpinner(-1, mpb.SpinnerOnLeft,
		mpb.PrependDecorators(decor.Name("processing ")),
		mpb.AppendDecorators(decor.Current(0, "%d items", decor.WC{W: 4})),
	)
	return &progress{p: p, bar: bar}
}

// stage returns a pass-through stage that ticks the spinner per item.
func (p *progress) stage() output.Stage {
	return Tap(func(jsondoc.JSONDoc) error {
		p.bar.Increment()
		return nil
	})
}

// done completes the spinner and flushes the final frame.
func (p *progress) done() {
	p.bar.SetTotal(p.bar.Current(), true)
	p.p.Wait()
}
