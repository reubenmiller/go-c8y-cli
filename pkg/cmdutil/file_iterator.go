package cmdutil

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
)

type FileIteratorOptions struct {
	Infinite        bool
	FirstNRows      int64
	Times           int64
	TotalRows       int64
	RandomSkip      float32
	DelayBefore     time.Duration
	Delay           time.Duration
	RandomDelayFunc flags.DurationGenerator
	Format          func(string, int64) string
}

func ExecuteFileIterator(w io.Writer, log *logger.Logger, files []string, iterFactory func(string) (iterator.Iterator, error), opt FileIteratorOptions) error {
	totalRows := opt.FirstNRows
	row := int64(0)
	rowCount := int64(0)
	outputCount := int64(1)

	if opt.RandomDelayFunc == nil {
		opt.RandomDelayFunc = func(d time.Duration) time.Duration { return opt.Delay }
	}

	outputFormatter := func(v string, rowNum int64) string {
		return v
	}
	if opt.Format != nil {
		outputFormatter = opt.Format
	}

	for {
		row++

		if row > opt.Times && !opt.Infinite {
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

				fmt.Fprintln(w, outputFormatter(string(responseText), outputCount))

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
