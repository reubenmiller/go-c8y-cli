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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
	"github.com/tidwall/pretty"
)

// DefaultSampleSize maximum number of rows sampled when resolving table columns
const DefaultSampleSize = 5

// DefaultTableFlushTimeout maximum time the first table rows are buffered while
// waiting for more rows to be sampled, so long-running streams (e.g. realtime
// subscriptions) are not delayed indefinitely
const DefaultTableFlushTimeout = 500 * time.Millisecond

// DefaultTableMaxWidth table width used when the terminal width can not be detected
const DefaultTableMaxWidth = 120

// Console thread safe way to write to an output
type Console struct {
	mu            sync.Mutex
	count         uint64
	out           io.Writer
	header        func([]string) []byte
	samples       []string
	sampleGroups  [][]jsonfilter.KeyGroup
	sampleCount   int
	sampleSize    int
	sampleTimeout time.Duration
	flushTimer    *time.Timer
	table         *encode.Table
	tableOptions  TableOptions
	Colorized     bool
	Compact       bool
	Disabled      bool
	Format        config.OutputFormat
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
	opts := TableOptions{
		MinColumnWidth: 2,
		MaxColumnWidth: 80,
		ColumnPadding:  15,
		SampleSize:     DefaultSampleSize,
		SampleTimeout:  DefaultTableFlushTimeout,
	}
	if tableOptions != nil {
		opts = *tableOptions
		if opts.SampleSize <= 0 {
			opts.SampleSize = DefaultSampleSize
		}
	}

	return &Console{
		out:           w,
		header:        header,
		sampleSize:    opts.SampleSize,
		sampleTimeout: opts.SampleTimeout,
		tableOptions:  opts,
		Format:        config.OutputTable,
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

// tableEngine returns the streaming table engine, creating it on first use.
// The caller must hold the mutex
func (c *Console) tableEngine() *encode.Table {
	if c.table == nil {
		maxColumnWidth := c.tableOptions.MaxColumnWidth
		if maxColumnWidth <= 0 {
			// no column width limit
			maxColumnWidth = 1 << 20
		}
		c.table = encode.NewTableWithWriter(
			tableviewer.NewStreamWriter(c.out, c.Colorized),
			encode.TableOptions{
				ColumnResolver:      c.resolveTableColumns,
				SampleSize:          c.sampleLimit(),
				MinColumnWidth:      c.tableOptions.MinColumnWidth,
				MinEmptyColumnWidth: c.tableOptions.MinEmptyValueColumnWidth,
				MaxColumnWidth:      maxColumnWidth,
				ColumnPadding:       c.tableOptions.ColumnPadding,
				MaxTableWidth:       tableviewer.TerminalWidth(DefaultTableMaxWidth),
				Formatter:           tableviewer.CellFormatter(c.tableOptions.NumberFormatter),
				Transform:           tableviewer.CellTransform(c.tableOptions.RowMode),
			})
	}
	return c.table
}

// resolveTableColumns resolves the table columns from the header samples
// collected via SetHeaderFromInput. It is invoked by the table engine from
// Write/Flush, so the caller already holds the mutex
func (c *Console) resolveTableColumns() []string {
	if cols := jsonfilter.MergeKeyGroups(c.sampleGroups); len(cols) > 0 {
		return cols
	}
	if len(c.samples) > 0 {
		return strings.Split(c.samples[0], ",")
	}
	return nil
}

// flushSample renders any rows buffered by the table engine's sample window
// without closing the table, so streamed output is not delayed indefinitely
func (c *Console) flushSample() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.table != nil {
		_ = c.table.Flush()
	}
}

// Flush renders any buffered table output and finalizes the table. It should
// be called once no more output is expected (e.g. at the end of a command)
func (c *Console) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.flushTimer != nil {
		c.flushTimer.Stop()
		c.flushTimer = nil
	}
	if c.table != nil {
		_ = c.table.Close()
		c.table = nil
	}
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

	if bt := bytes.TrimSpace(b); c.IsJSON() && (jsonUtilities.IsJSONArray(bt) || jsonUtilities.IsJSONObject(bt)) {

		switch c.Format {
		case config.OutputTable:
			table := c.tableEngine()
			// The table engine buffers documents while sampling, so the
			// rows must not reference the caller's buffer
			doc := jsondoc.New(bytes.Clone(bt))
			for d := range doc.Iter() {
				if err := table.Write(d); err != nil {
					return 0, err
				}
			}
			// Don't delay streamed output indefinitely (e.g. realtime subscriptions)
			if c.flushTimer == nil && c.sampleTimeout > 0 {
				c.flushTimer = time.AfterFunc(c.sampleTimeout, c.flushSample)
			}
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
