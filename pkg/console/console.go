package console

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
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
	IsJSON      bool
}

// NewConsole create a new console writter
func NewConsole(w io.Writer, header func([]string) []byte) *Console {
	return &Console{
		out:    w,
		header: header,
	}
}

func (c *Console) SetHeaderFromInput(input string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sampleCount < 5 {
		c.samples = append(c.samples, input)
		c.sampleCount++
	}
	return
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
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.count == 0 && c.header != nil {
		c.out.Write(c.header(c.samples))
	}
	c.count++

	if bt := bytes.TrimSpace(b); c.IsJSON && (jsonUtilities.IsJSONArray(bt) || jsonUtilities.IsJSONObject(bt)) {
		b = pretty.PrettyOptions(b, &pretty.Options{
			SortKeys: true,
			Width:    80,
			Prefix:   "",
			Indent:   "  ",
		})
		if c.Compact {
			b = append(pretty.Ugly(b), '\n')
		}
		if c.Colorized {
			b = pretty.Color(b, pretty.TerminalStyle)
		}
	}

	return c.out.Write(b)
}
