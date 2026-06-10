package console

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/numbers"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/tableviewer"
	"github.com/tidwall/pretty"
)

// DefaultSampleSize maximum number of rows sampled when resolving table columns
const DefaultSampleSize = 5

// DefaultTableFlushTimeout maximum time the first table rows are buffered while
// waiting for more rows to be sampled, so long-running streams (e.g. realtime
// subscriptions) are not delayed indefinitely
const DefaultTableFlushTimeout = 500 * time.Millisecond

// Console thread safe way to write to an output
type Console struct {
	mu               sync.Mutex
	count            uint64
	out              io.Writer
	header           func([]string) []byte
	samples          []string
	sampleGroups     [][]jsonfilter.KeyGroup
	sampleCount      int
	sampleSize       int
	sampleTimeout    time.Duration
	tableBuffer      [][]byte
	tableFlushed     bool
	tableHeaderShown bool
	flushTimer       *time.Timer
	Colorized        bool
	Compact          bool
	Disabled         bool
	Format           config.OutputFormat
	TableViewer      *tableviewer.TableView
}

// TableOptions table options to control the column behaviour
type TableOptions struct {
	// MinColumnWidth minimum column width
	MinColumnWidth int

	// MinEmptyValueColumnWidth minimum column width to use when the value is empty
	// If set to 0, then the MinColumnWidth will be used
	MinEmptyValueColumnWidth int

	// MaxColumnWidth maximum column width
	MaxColumnWidth int

	// ColumnPadding column padding
	ColumnPadding int

	// Row mode (truncate or wrap)
	RowMode string

	// NumberFormatter formatting used when rendering numbers
	NumberFormatter numbers.NumberFormatter

	// SampleSize maximum number of rows which are sampled when resolving
	// the table columns and column widths. If set to 0, then a default value will be used
	SampleSize int

	// SampleTimeout maximum duration to buffer rows whilst waiting for more rows
	// to be sampled, so that streamed output is not delayed indefinitely.
	// If set to 0, then buffered rows are only rendered once the sample size is
	// reached or no more output is expected
	SampleTimeout time.Duration
}

// NewConsole create a new console writer
func NewConsole(w io.Writer, tableOptions *TableOptions, header func([]string) []byte) *Console {
	minColumnWidth := 2
	maxColumnWidth := 80
	columnPadding := 15
	minEmptyWidth := 0
	rowMode := ""
	sampleSize := DefaultSampleSize
	sampleTimeout := DefaultTableFlushTimeout
	var numberFormatter numbers.NumberFormatter

	if tableOptions != nil {
		minColumnWidth = tableOptions.MinColumnWidth
		maxColumnWidth = tableOptions.MaxColumnWidth
		columnPadding = tableOptions.ColumnPadding
		minEmptyWidth = tableOptions.MinEmptyValueColumnWidth
		rowMode = tableOptions.RowMode
		numberFormatter = tableOptions.NumberFormatter
		if tableOptions.SampleSize > 0 {
			sampleSize = tableOptions.SampleSize
		}
		sampleTimeout = tableOptions.SampleTimeout
	}

	return &Console{
		out:           w,
		header:        header,
		sampleSize:    sampleSize,
		sampleTimeout: sampleTimeout,
		TableViewer: &tableviewer.TableView{
			Out:                      w,
			MinColumnWidth:           minColumnWidth,
			MaxColumnWidth:           maxColumnWidth,
			ColumnPadding:            columnPadding,
			MinEmptyValueColumnWidth: minEmptyWidth,
			EnableColor:              false,
			RowMode:                  rowMode,
			NumberFormatter:          numberFormatter,
			SampleSize:               sampleSize,
		},
		Format: config.OutputTable,
	}
}

func (c *Console) SetOut(w io.Writer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.out = w
}

// IsCSV return true if csv output is set
func (c *Console) IsCSV() bool {
	return c.Format == config.OutputCSV || c.Format == config.OutputCSVWithHeader
}

func (c *Console) IsTextOutput() bool {
	return c.Format == config.OutputCSV || c.Format == config.OutputCSVWithHeader || c.Format == config.OutputTSV || c.Format == config.OutputCompletion
}

// WithCSVHeader returns true if the csv output should include a header
func (c *Console) WithCSVHeader() bool {
	return c.Format == config.OutputCSVWithHeader
}

// IsJSON return true if JSON output is set
func (c *Console) IsJSON() bool {
	return c.Format != config.OutputCSV && c.Format != config.OutputCSVWithHeader
}

// IsJSONStream check if json stream mode is activated
func (c *Console) IsJSONStream() bool {
	// deprecated
	return false
}

// IsTable return true if table output is set
func (c *Console) IsTable() bool {
	return c.Format == config.OutputTable
}

func (c *Console) SetHeaderFromInput(input string, groups []jsonfilter.KeyGroup) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sampleCount < c.sampleLimit() {
		c.samples = append(c.samples, input)
		c.sampleGroups = append(c.sampleGroups, groups)
		c.sampleCount++
	}
}

// sampleLimit maximum number of rows which are sampled when resolving table columns
func (c *Console) sampleLimit() int {
	if c.sampleSize > 0 {
		return c.sampleSize
	}
	return DefaultSampleSize
}

// renderTableRow render a single row, only including the table header if it has not been shown yet
func (c *Console) renderTableRow(b []byte) {
	c.TableViewer.Render(b, !c.tableHeaderShown)
	c.tableHeaderShown = true
}

// flushTable resolve the table columns from the sampled rows and render any buffered rows.
// The caller must hold the mutex
func (c *Console) flushTable() {
	if c.tableFlushed {
		return
	}
	c.tableFlushed = true
	if c.flushTimer != nil {
		c.flushTimer.Stop()
		c.flushTimer = nil
	}

	if len(c.TableViewer.Columns) == 0 {
		if cols := jsonfilter.MergeKeyGroups(c.sampleGroups); len(cols) > 0 {
			c.TableViewer.Columns = cols
		} else if len(c.samples) > 0 {
			c.TableViewer.Columns = strings.Split(c.samples[0], ",")
		}
	}

	// resolve the column widths from all of the sampled rows (not just the first row)
	c.TableViewer.SampleColumnWidths(c.tableBuffer)

	for _, row := range c.tableBuffer {
		c.renderTableRow(row)
	}
	c.tableBuffer = nil
}

// Flush render any buffered table output. It should be called once
// no more output is expected (e.g. at the end of a command)
func (c *Console) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flushTable()
}

// Printf mimics fmt.Printf
func (c *Console) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(c, format, a...)
}

// Println mimics fmt.Println
func (c *Console) Println(a ...interface{}) (int, error) {
	return fmt.Fprintln(c, a...)
}

// Write a line to the output. Supports concurrent access
func (c *Console) Write(b []byte) (n int, err error) {
	if c.Disabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.count == 0 && c.header != nil {
		fmt.Fprintf(c.out, "%s", c.header(c.samples))
	}
	c.count++
	c.TableViewer.EnableColor = c.Colorized

	if bt := bytes.TrimSpace(b); c.IsJSON() && (jsonUtilities.IsJSONArray(bt) || jsonUtilities.IsJSONObject(bt)) {

		switch c.Format {
		case config.OutputTable:
			if !c.tableFlushed {
				// Buffer the first rows so the table columns can be resolved
				// against multiple rows instead of just the first one, as the
				// first row may not contain all of the selected fragments
				row := make([]byte, len(b))
				copy(row, b)
				c.tableBuffer = append(c.tableBuffer, row)
				if len(c.tableBuffer) >= c.sampleLimit() {
					c.flushTable()
				} else if c.flushTimer == nil && c.sampleTimeout > 0 {
					// don't delay streamed output indefinitely (e.g. realtime subscriptions)
					c.flushTimer = time.AfterFunc(c.sampleTimeout, func() {
						c.Flush()
					})
				}
				return 0, nil
			}
			c.renderTableRow(b)
			return 0, nil
		}

		b = pretty.PrettyOptions(b, &pretty.Options{
			SortKeys: true,
			Width:    80,
			Prefix:   "",
			Indent:   "  ",
		})
		if c.Compact || c.IsCSV() {
			b = append(pretty.Ugly(b), '\n')
		}
		if c.Colorized && c.IsJSON() {
			b = pretty.Color(b, pretty.TerminalStyle)
		}
	}

	return c.out.Write(b)
}
