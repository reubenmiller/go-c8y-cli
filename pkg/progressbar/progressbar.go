package progressbar

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

// ProgressBar Multi progress bar
type ProgressBar struct {
	mu                   sync.Mutex
	p                    *mpb.Progress
	Bars                 []*mpb.Bar
	TaskName             string
	NumBars              int
	total                int
	enabled              bool
	started              bool
	overviewTotal        int64
	overviewCurrentTotal int64
	refreshRate          time.Duration
	w                    io.Writer
}

// NewMultiProgressBar create a new progress bar to display progress of workers and an overall progress
func NewMultiProgressBar(w io.Writer, total, numBars int, name string, enable bool) *ProgressBar {
	var p *mpb.Progress
	refreshRate := 120 * time.Millisecond
	if enable {
		p = mpb.New(
			mpb.ContainerOptional(
				mpb.WithRefreshRate(refreshRate), true),
			mpb.WithOutput(w),
		)
	}

	if name != "" {
		name = "sending request"
	}

	return &ProgressBar{
		p:           p,
		NumBars:     numBars,
		TaskName:    name,
		total:       total,
		enabled:     enable,
		refreshRate: refreshRate,
		w:           w,
	}
}

// IsEnabled check if the progress bar is enabled or not.
func (p *ProgressBar) IsEnabled() bool {
	return p.enabled
}

// IsRunning check if the progress bar is running or not (aka. has started)
func (p *ProgressBar) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.started
}

// RefreshRate returns the configured refresh rate of the progress bar
func (p *ProgressBar) RefreshRate() time.Duration {
	return p.refreshRate
}

// IncrementOverviewTotal increment the current overview value by 1
func (p *ProgressBar) IncrementOverviewCurrent() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.overviewCurrentTotal++
	if len(p.Bars) > 0 {
		p.Bars[0].SetCurrent(p.overviewCurrentTotal)
	}
}

// IncrementOverviewTotal increment the overview total by 1
func (p *ProgressBar) IncrementOverviewTotal() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.overviewTotal++
	if len(p.Bars) > 0 {
		p.Bars[0].SetTotal(p.overviewTotal, false)
	}
}

// IncrementWorkerTotal increments the worker total by 1
func (p *ProgressBar) IncrementWorkerTotal(w int, total int64) {
	if w <= len(p.Bars)-1 {
		p.Bars[w].SetTotal(total+1, false)
	}
	p.IncrementOverviewTotal()
}

// StartJob starts a job for a given worker. It will record the time it started the job, and the duration will be used in calculated once FinishedJob is called.
func (p *ProgressBar) StartJob(w int, total int64) time.Time {
	if p.IsEnabled() {
		if w > 0 && w <= p.NumBars {
			p.IncrementWorkerTotal(w, total+1)
		}
	}
	return time.Now()
}

// FinishedJob finish a job and record the time since the job started (if StartJob was called)
func (p *ProgressBar) FinishedJob(w int, startTime time.Time) {
	if p.IsEnabled() {
		p.IncrementOverviewCurrent()

		if p.IsRunning() {
			p.Bars[w].Increment()
			duration := time.Since(startTime)
			p.Bars[w].DecoratorEwmaUpdate(duration)
		}
	}
}

// WorkerCompleted marks a worker as completed. The current total will be used as its final value
func (p *ProgressBar) WorkerCompleted(w int) {
	if p.IsEnabled() {
		if p.IsRunning() {
			p.Bars[w].SetTotal(0, true)
		}
	}
}

// Completed set overall completed
func (p *ProgressBar) Completed() {
	if p.IsEnabled() {
		if p.IsRunning() {
			p.Bars[0].SetTotal(0, true)
		}
	}
}

// Start start displaying the progress bar
func (p *ProgressBar) Start(age float64) {

	if !p.IsEnabled() {
		return
	}

	if age <= 0 {
		age = 5
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return
	}

	// add new line before progress for a cleaner look
	fmt.Fprintln(p.w, "")

	overviewLabel := "(started: " + time.Now().Format(time.RFC3339) + ")"
	overviewBar := p.p.AddSpinner(int64(p.total), mpb.SpinnerOnRight,
		mpb.PrependDecorators(
			decor.Name("elapsed", decor.WC{W: len("elapsed") + 1, C: decor.DidentRight}),
			decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
			decor.Name(overviewLabel, decor.WC{W: len(overviewLabel) + 1, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.Current(0, "total requests sent:  %d", decor.WC{W: 8}),
		),
	)
	if p.overviewCurrentTotal > 0 {
		// set initial value (requires setting total then current!)
		overviewBar.SetTotal(p.overviewCurrentTotal+1, false)
		overviewBar.SetCurrent(p.overviewCurrentTotal)
	}
	p.Bars = append(p.Bars, overviewBar)

	for i := 0; i < p.NumBars; i++ {
		worker := fmt.Sprintf("worker %d:", i+1)
		bar := p.p.AddSpinner(int64(p.total), mpb.SpinnerOnLeft,
			mpb.PrependDecorators(
				decor.Name(worker, decor.WC{W: len(worker) + 5, C: decor.DidentRight}),
			),
			mpb.AppendDecorators(
				decor.Name(" avg: "),
				decor.EwmaSpeed(0, "%.3f request/s", age),
				// decor.OnComplete(decor.EwmaSpeed(0, "%.3f request/s", age), ""),
			),
		)
		p.Bars = append(p.Bars, bar)
	}
	p.started = true
}

// Wait waits for the progress bar to finish
func (p *ProgressBar) Wait() {
	if p.IsEnabled() && p.IsRunning() {
		p.Wait()
	}
}
