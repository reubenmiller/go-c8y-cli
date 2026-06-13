package c8ystream

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
)

// ErrorCollector is a stage that records item-level errors and lets the
// stream continue, replicating the CLI batch semantics where one failed
// input does not abort the remaining inputs. Place it before the other
// stages so filters and templates only see successful documents.
//
// Max > 0 aborts the stream once that many errors have been recorded
// (the --abortOnErrors behaviour). The zero value never aborts.
//
// Stages run on the consuming goroutine, so no locking is required even when
// jobs execute concurrently.
type ErrorCollector struct {
	Errors []error
	Max    int
}

// Stage returns the stream stage that performs the collection.
func (c *ErrorCollector) Stage() output.Stage {
	return func(src output.Seq) output.Seq {
		return func(yield func(jsondoc.JSONDoc, error) bool) {
			for doc, err := range src {
				if err == nil {
					if !yield(doc, nil) {
						return
					}
					continue
				}
				c.Errors = append(c.Errors, err)
				if c.Max > 0 && len(c.Errors) >= c.Max {
					yield(jsondoc.Empty(), fmt.Errorf("aborted batch as the error count (%d) reached the limit: %w", len(c.Errors), err))
					return
				}
			}
		}
	}
}

// Err returns the aggregate result for use as the command error, so the CLI
// exits non-zero when any input item failed.
func (c *ErrorCollector) Err() error {
	switch len(c.Errors) {
	case 0:
		return nil
	case 1:
		return c.Errors[0]
	default:
		return fmt.Errorf("batch completed with %d errors, first error: %w", len(c.Errors), c.Errors[0])
	}
}

// Tap returns a stage that passes every document through unchanged while
// also handing it to fn — teeing the stream at that point (e.g. writing a
// backup of processed items while they continue to the terminal renderer).
func Tap(fn func(jsondoc.JSONDoc) error) output.Stage {
	return output.Map(func(doc jsondoc.JSONDoc) (jsondoc.JSONDoc, error) {
		if err := fn(doc); err != nil {
			return doc, err
		}
		return doc, nil
	})
}
