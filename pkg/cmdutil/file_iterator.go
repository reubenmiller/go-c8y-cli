package cmdutil

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/randdata"
)

type RandomRange struct {
	Min *int64
	Max *int64
}

type FileIteratorOptions struct {
	Infinite        bool
	FirstNRows      int64
	Times           RandomRange
	TotalRows       int64
	RandomSkip      float32
	DelayBefore     time.Duration
	Delay           time.Duration
	RandomDelayFunc flags.DurationGenerator
	OutputFunc      func([]byte) error
	Format          func([]byte, int64) []byte
}

func ExecuteFileIterator(w io.Writer, log *logger.Logger, files []string, iterFactory func(string) (iterator.Iterator, error), opt FileIteratorOptions) error {
	totalRows := opt.FirstNRows
	row := int64(0)
	rowCount := int64(0)
	outputCount := int64(1)

	if opt.RandomDelayFunc == nil {
		opt.RandomDelayFunc = func(d time.Duration) time.Duration { return opt.Delay }
	}

	outputFormatter := func(v []byte, rowNum int64) []byte {
		return v
	}
	if opt.Format != nil {
		outputFormatter = opt.Format
	}

	var times int64

	// randomized times
	if opt.Times.Min != nil || opt.Times.Max != nil {
		// If only min is provided, then adjust max values to equal to min
		// This will behaviour exactly the same as using --times x.
		// However, just providing a max value will result in range from 1 to max
		if opt.Times.Max == nil {
			opt.Times.Max = opt.Times.Min
		}
		// Allow users to set --max 0 to disable all output
		// as it gives the user full control to also turn off the output if desired
		if *opt.Times.Max == 0 {
			times = 0
		} else {
			times = randdata.Integer(*opt.Times.Max, *opt.Times.Min)
		}
	}

	for {
		row++

		if row > times && !opt.Infinite {
			break
		}

		for _, curFile := range files {
			iter, err := iterFactory(curFile)
			if err != nil {
				return err
			}

			for {
				responseText, _, err := iter.GetNext()
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}

				if totalRows != 0 && rowCount >= totalRows {
					log.Debugf("Found first %d rows", rowCount)
					return nil
				}

				if opt.RandomSkip >= -1 {
					// randomly skip a row. 1 = always skip, 0 = never skip
					randValue := rand.Float32()
					if randValue <= opt.RandomSkip {
						log.Debugf("Skipping random row: %d. value=%f, limit=%f", row, randValue, opt.RandomSkip)
						continue
					}
				}

				if opt.DelayBefore > 0 {
					time.Sleep(opt.DelayBefore)
				}

				if opt.OutputFunc != nil {
					if err := opt.OutputFunc(outputFormatter(responseText, outputCount)); err != nil {
						return err
					}
				} else {
					fmt.Fprintln(w, outputFormatter(responseText, outputCount))
				}

				currentDelay := opt.RandomDelayFunc(opt.Delay)
				if currentDelay > 0 {
					log.Infof("Waiting %v before printing next value", currentDelay)
					time.Sleep(currentDelay)
				}
				outputCount++
				rowCount++
			}
		}
	}
	return nil
}
