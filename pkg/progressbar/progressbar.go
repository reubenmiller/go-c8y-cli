package progressbar

import (
	"fmt"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

type ProgressBar struct {
	mu            sync.Mutex
	p             *mpb.Progress
	Bars          []*mpb.Bar
	TaskName      string
	NumBars       int
	total         int
	enabled       bool
	overviewTotal int64
}

func NewMulitProgressBar(total, numBars int, name string, enable bool) *ProgressBar {
	var p *mpb.Progress
	if enable {
		p = mpb.New(
			mpb.ContainerOptional(
				mpb.WithRefreshRate(120*time.Millisecond), true),
		)
	}

	if name != "" {
		name = "sending request"
	}

	return &ProgressBar{
		p:        p,
		NumBars:  numBars,
		TaskName: name,
		total:    total,
		enabled:  enable,
	}
}

// func (p *ProgressBar) Render(s *mpb.Progress) string {
// 	str := fmt.Sprintf("%d/%d", s.Current, s.Total)
// 	return fmt.Sprintf("%8s", str)

// 	decor := func(s *mpb.Statistics) string {

// 	}
// }

func (p *ProgressBar) IsEnabled() bool {
	return p.enabled
}

func (p *ProgressBar) IncrementOverviewCurrent() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Bars[0].Increment()
}

func (p *ProgressBar) IncrementOverviewTotal() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.overviewTotal++
	p.Bars[0].SetTotal(p.overviewTotal, false)
}

func (p *ProgressBar) StartJob(w int, total int64) time.Time {
	if p.IsEnabled() {
		if w > 0 && w <= p.NumBars {
			p.Bars[w].SetTotal(total+1, false)
			p.IncrementOverviewTotal()
		}
	}
	return time.Now()
}

func (p *ProgressBar) FinishedJob(w int, startTime time.Time) {
	if p.IsEnabled() {
		p.IncrementOverviewCurrent()

		p.Bars[w].Increment()
		duration := time.Since(startTime)
		p.Bars[w].DecoratorEwmaUpdate(duration)
	}
}

func (p *ProgressBar) WorkerCompleted(w int) {
	if p.IsEnabled() {
		p.Bars[w].Completed()
	}
}

func (p *ProgressBar) Completed() {
	if p.IsEnabled() {
		p.Bars[0].Completed()
	}
}

func (p *ProgressBar) Start(age float64) {

	if !p.IsEnabled() {
		return
	}

	if age <= 0 {
		age = 5
	}

	overviewLabel := "total"
	overviewBar := p.p.AddSpinner(int64(p.total), mpb.SpinnerOnLeft,
		mpb.PrependDecorators(
			decor.Elapsed(decor.ET_STYLE_MMSS, decor.WC{W: 8, C: decor.DidentRight}),
			decor.Name(overviewLabel, decor.WC{W: len(overviewLabel) + 1, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.Current(0, "%d", decor.WC{W: 5}),
		),
	)

	p.Bars = append(p.Bars, overviewBar)

	for i := 0; i < p.NumBars; i++ {
		worker := fmt.Sprintf("worker %d:", i+1)
		bar := p.p.AddSpinner(int64(p.total), mpb.SpinnerOnLeft,
			mpb.PrependDecorators(
				decor.Name(worker, decor.WC{W: len(worker) + 5, C: decor.DidentRight}),
			),
			mpb.AppendDecorators(
				decor.Name(" avg: "),
				decor.OnComplete(decor.EwmaSpeed(0, "%.3f request/s", age), "done"),
			),
		)
		p.Bars = append(p.Bars, bar)
	}
	return
}

func (p *ProgressBar) Wait() {
	if p.IsEnabled() {
		p.Wait()
	}
}
