package iostreams

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/vbauerster/mpb/v6"
)

func newProgressTestStreams() *IOStreams {
	s, _, _, _ := Test()
	s.ErrOut = io.Discard
	s.SetProgress(true)
	return s
}

// simulateDownload mimics the request handler: a bar is added to the shared progress
// indicator when the response is received, the body is read, and then
// the handler waits for the progress indicator before printing the output
func simulateDownload(s *IOStreams) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	const size = 10
	bar := s.ProgressIndicator().Add(size, mpb.NewBarFiller("[=>-]"))
	bar.IncrInt64(size)
	s.WaitForProgressIndicator()
	return nil
}

// Issue #711: progress indicator should be usable again after it has been waited on
func Test_ProgressIndicatorIsReusableAfterWait(t *testing.T) {
	s := newProgressTestStreams()
	for i := range 3 {
		if err := simulateDownload(s); err != nil {
			t.Fatalf("download %d failed: %s", i+1, err)
		}
	}
}

// Multiple workers (e.g. --workers 5) share the same progress indicator
func Test_ProgressIndicatorConcurrentWorkers(t *testing.T) {
	const workers = 5
	const downloadsPerWorker = 2000

	s := newProgressTestStreams()

	var wg sync.WaitGroup
	var failures atomic.Int64
	var firstErr atomic.Value
	for range workers {
		wg.Go(func() {
			for range downloadsPerWorker {
				if err := simulateDownload(s); err != nil {
					failures.Add(1)
					firstErr.CompareAndSwap(nil, err)
				}
			}
		})
	}
	wg.Wait()

	if n := failures.Load(); n > 0 {
		t.Fatalf("%d/%d downloads panicked. first error: %v", n, workers*downloadsPerWorker, firstErr.Load())
	}
}
