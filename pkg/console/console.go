package console

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/numbers"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/tableviewer"
	"github.com/tidwall/pretty"
)

// Console thread safe way to write to an output
type Console struct {
	mu          sync.Mutex
	count       uint64
	out         io.Writer
	header      func([]string) []byte
	samples     []string
	sampleCount int
	Colorized   bool
	Compact     bool
	Disabled    bool
	Format      config.OutputFormat
	TableViewer *tableviewer.TableView
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
}

// NewConsole create a new console writer
func NewConsole(w io.Writer, tableOptions *TableOptions, header func([]string) []byte) *Console {
	minColumnWidth := 2
	maxColumnWidth := 80
	columnPadding := 15
	minEmptyWidth := 0
	rowMode := ""
	var numberFormatter numbers.NumberFormatter

	if tableOptions != nil {
		minColumnWidth = tableOptions.MinColumnWidth
		maxColumnWidth = tableOptions.MaxColumnWidth
		columnPadding = tableOptions.ColumnPadding
		minEmptyWidth = tableOptions.MinEmptyValueColumnWidth
		rowMode = tableOptions.RowMode
		numberFormatter = tableOptions.NumberFormatter
	}

	return &Console{
		out:    w,
		header: header,
		TableViewer: &tableviewer.TableView{
			Out:                      w,
			MinColumnWidth:           minColumnWidth,
			MaxColumnWidth:           maxColumnWidth,
			ColumnPadding:            columnPadding,
			MinEmptyValueColumnWidth: minEmptyWidth,
			EnableColor:              false,
			RowMode:                  rowMode,
			NumberFormatter:          numberFormatter,
		},
		Format: config.OutputTable,
	}
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

func (c *Console) SetHeaderFromInput(input string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sampleCount < 5 {
		c.samples = append(c.samples, input)
		c.sampleCount++
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
	showHeader := c.count == 0
	c.count++
	c.TableViewer.EnableColor = c.Colorized

	if bt := bytes.TrimSpace(b); c.IsJSON() && (jsonUtilities.IsJSONArray(bt) || jsonUtilities.IsJSONObject(bt)) {

		switch c.Format {
		case config.OutputTable:
			cols := []string{}
			if len(c.samples) > 0 {
				cols = append(cols, strings.Split(c.samples[0], ",")...)
			}
			if len(c.TableViewer.Columns) == 0 {
				c.TableViewer.Columns = cols
			}
			c.TableViewer.Render(b, showHeader)
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
