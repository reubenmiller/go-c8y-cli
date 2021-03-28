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
	valueCh := make(chan (interface{}))

	if interval <= 0 {
		interval = 1000 * time.Millisecond
	}
	if timeout <= 0 {
		timeout = 300 * time.Second
	}

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	timeoutCh := time.After(timeout)

	go func() {
		for {
			item, err := predicate.Get()
			if err != nil {

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
			return lastValue, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitTimeout, fmt.Sprintf("Timeout after %d second/s", int64(timeout/time.Second)))

		case msg := <-valueCh:
			if msg != nil {
				lastValue = msg
			}
			done, err := predicate.Check(msg)
			if done {
				return msg, err
			}

		case <-signalCh:
			// Enable ctrl-c to stop
			return lastValue, cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCancel, "User cancelled command")
		}
	}
}
