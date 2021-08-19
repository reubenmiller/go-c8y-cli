package desiredstate

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
)

type StateDefiner interface {
	Get() (interface{}, error)
	Check(interface{}) (bool, error)
}

func WaitFor(interval time.Duration, timeout time.Duration, predicate StateDefiner) (interface{}, error) {
	return wait(-1, interval, timeout, predicate)
}

// WaitForWithRetries wait for a predicate to be true and limiting the retries by an explict count
func WaitForWithRetries(retries int64, interval time.Duration, timeout time.Duration, predicate StateDefiner) (interface{}, error) {
	return wait(retries, interval, timeout, predicate)
}

func wait(retries int64, interval time.Duration, timeout time.Duration, predicate StateDefiner) (interface{}, error) {
	valueCh := make(chan (interface{}))
	var attemptCounter int64

	if interval <= 0 {
		interval = 1000 * time.Millisecond
	}
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	timeoutCh := make(<-chan time.Time)
	if timeout > 0 {
		timeoutCh = time.After(timeout)
	}

	go func() {
		for {
			item, err := predicate.Get()
			if err != nil {
				valueCh <- err
			} else {
				valueCh <- item
			}
			time.Sleep(interval)
		}
	}()

	var lastValue interface{}
	for {
		select {
		case <-timeoutCh:
			return lastValue, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitTimeout, fmt.Sprintf("Timeout after %v", timeout))

		case msg := <-valueCh:
			attemptCounter++
			if err, ok := msg.(error); ok {
				return nil, err
			}
			if msg != nil {
				lastValue = msg
			}
			done, err := predicate.Check(msg)
			if done {
				return msg, err
			}

			if retries >= 0 && attemptCounter > retries {
				return msg, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCancel, "Max retries exceeded")
			}

		case <-signalCh:
			// Enable ctrl-c to stop
			return lastValue, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCancel, "User cancelled command")
		}
	}
}
