package console

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/tableviewer"
	"github.com/tidwall/pretty"
)

type OutputFormat int

const (
	// OutputJSON json output
	OutputJSON OutputFormat = iota

	// OutputTable table output
	OutputTable

	// OutputCSV csv output
	OutputCSV

	// OutputCSVWithHeader csv output with header
	OutputCSVWithHeader
)

func (f OutputFormat) String() string {
	values := map[OutputFormat]string{
		OutputJSON:          "json",
		OutputTable:         "table",
		OutputCSV:           "csv",
		OutputCSVWithHeader: "csvheader",
	}

	if v, ok := values[f]; ok {
		return v
	}
	return ""
}

func (f OutputFormat) FromString(name string) OutputFormat {
	values := map[string]OutputFormat{
		"json":      OutputJSON,
		"table":     OutputTable,
		"csv":       OutputCSV,
		"csvheader": OutputCSVWithHeader,
	}

	if v, ok := values[strings.ToLower(name)]; ok {
		return v
	}
	return f
}

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
	Format      OutputFormat
	TableViewer *tableviewer.TableView
}

// NewConsole create a new console writter
func NewConsole(w io.Writer, header func([]string) []byte) *Console {
	return &Console{
		out:    w,
		header: header,
		TableViewer: &tableviewer.TableView{
			Out:            w,
			MinColumnWidth: 2,
			MaxColumnWidth: 80,
			ColumnPadding:  5,
			EnableColor:    false,
		},
		Format: OutputTable,
	}
}

// IsCSV return true if csv output is set
func (c *Console) IsCSV() bool {
	return c.Format == OutputCSV || c.Format == OutputCSVWithHeader
}

// WithCSVHeader returns true if the csv output should include a header
func (c *Console) WithCSVHeader() bool {
	return c.Format == OutputCSVWithHeader
}

// IsJSON return true if JSON output is set
func (c *Console) IsJSON() bool {
	return c.Format != OutputCSV && c.Format != OutputCSVWithHeader
}

// IsJSONStream check if json stream mode is activated
func (c *Console) IsJSONStream() bool {
	// deprecated
	return false
}

// IsTable return true if table output is set
func (c *Console) IsTable() bool {
	return c.Format == OutputTable
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
		case OutputTable:
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
