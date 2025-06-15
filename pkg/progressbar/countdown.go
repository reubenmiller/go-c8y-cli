package progressbar

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

// CountdownProgressBar progress bar to show a countdown
type CountdownProgressBar struct {
	mu          sync.Mutex
	p           *mpb.Progress
	spinner     *mpb.Bar
	TaskName    string
	StartedAt   time.Time
	ExpiresAt   time.Time
	Duration    time.Duration
	enabled     bool
	started     bool
	refreshRate time.Duration
	w           io.Writer
}

// NewCountdownProgressBar create a new progress bar to display countdown timer
func NewCountdownProgressBar(w io.Writer, duration time.Duration, name string, enable bool) *CountdownProgressBar {
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

	return &CountdownProgressBar{
		p:           p,
		TaskName:    name,
		Duration:    duration,
		enabled:     enable,
		refreshRate: refreshRate,
		w:           w,
	}
}

// IsEnabled check if the progress bar is enabled or not.
func (p *CountdownProgressBar) IsEnabled() bool {
	return p.enabled
}

func (p *CountdownProgressBar) UpdateValue(l string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	completed := time.Since(p.StartedAt).Seconds()
	p.spinner.SetCurrent(int64(completed))
}

// IsRunning check if the progress bar is running or not (aka. has started)
func (p *CountdownProgressBar) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.started
}

// RefreshRate returns the configured refresh rate of the progress bar
func (p *CountdownProgressBar) RefreshRate() time.Duration {
	return p.refreshRate
}

// Start start displaying the progress bar
func (p *CountdownProgressBar) Start() {

	if !p.IsEnabled() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return
	}

	// add new line before progress for a cleaner look
	fmt.Fprintln(p.w, "")
	p.StartedAt = time.Now()
	p.ExpiresAt = p.StartedAt.Add(p.Duration)

	overviewLabel := "(expires at: " + p.ExpiresAt.Format(time.RFC3339) + ")"
	spinner := p.p.AddSpinner(int64(p.Duration.Seconds()), mpb.SpinnerOnRight,
		mpb.PrependDecorators(
			decor.Name("remaining", decor.WC{W: len("remaining") + 1, C: decor.DidentRight}),
			Remaining(p.ExpiresAt, decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
			decor.Name(overviewLabel, decor.WC{W: len(overviewLabel) + 1, C: decor.DidentRight}),
		),
	)
	p.spinner = spinner
	p.started = true
}

// Wait waits for the progress bar to finish
func (p *CountdownProgressBar) Wait() {
	if p.IsEnabled() && p.IsRunning() {
		p.Wait()
	}
}

// Remaining decorator. It's wrapper of NewRemaining.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`wcc` optional WC config
func Remaining(endTime time.Time, style decor.TimeStyle, wcc ...decor.WC) decor.Decorator {
	return NewRemaining(style, endTime, wcc...)
}

// NewRemaining returns remaining time decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`endTime` end time
//
//	`wcc` optional WC config
func NewRemaining(style decor.TimeStyle, endTime time.Time, wcc ...decor.WC) decor.Decorator {
	var msg string
	producer := chooseTimeProducer(style)
	fn := func(s decor.Statistics) string {
		if !s.Completed {
			msg = producer(max(time.Until(endTime), 0))
		}
		return msg
	}
	return decor.Any(fn, wcc...)
}

func chooseTimeProducer(style decor.TimeStyle) func(time.Duration) string {
	switch style {
	case decor.ET_STYLE_HHMMSS:
		return func(remaining time.Duration) string {
			hours := int64(remaining/time.Hour) % 60
			minutes := int64(remaining/time.Minute) % 60
			seconds := int64(remaining/time.Second) % 60
			return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		}
	case decor.ET_STYLE_HHMM:
		return func(remaining time.Duration) string {
			hours := int64(remaining/time.Hour) % 60
			minutes := int64(remaining/time.Minute) % 60
			return fmt.Sprintf("%02d:%02d", hours, minutes)
		}
	case decor.ET_STYLE_MMSS:
		return func(remaining time.Duration) string {
			hours := int64(remaining/time.Hour) % 60
			minutes := int64(remaining/time.Minute) % 60
			seconds := int64(remaining/time.Second) % 60
			if hours > 0 {
				return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
			}
			return fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
	default:
		return func(remaining time.Duration) string {
			// strip off nanoseconds
			return ((remaining / time.Second) * time.Second).String()
		}
	}
}
