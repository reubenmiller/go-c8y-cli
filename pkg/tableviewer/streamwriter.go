package tableviewer

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
)

// StreamWriter adapts tablewriter's streaming renderer to the go-c8y/v2
// output table engine: the SDK side owns column resolution and width
// sampling (encode.Table) and hands resolved rows to this writer, which
// renders them incrementally — rows are written as they arrive instead of
// buffering the whole result set.
//
// Usage:
//
//	err := output.Render(ctx, src,
//	    encode.NewTableWithWriter(tableviewer.NewStreamWriter(os.Stdout, false), encode.TableOptions{
//	        Columns: []string{"id", "name"},
//	    }))
type StreamWriter struct {
	out         io.Writer
	enableColor bool
	table       *tablewriter.Table
}

var _ encode.RowWriter = (*StreamWriter)(nil)

func NewStreamWriter(out io.Writer, enableColor bool) *StreamWriter {
	return &StreamWriter{out: out, enableColor: enableColor}
}

// WriteHeader receives the resolved columns and sampled widths from the
// table engine and starts the underlying streaming table. Fixed per-column
// widths are required for streaming output (rows rendered early cannot be
// re-aligned), which is exactly what the engine's sampling provides.
func (s *StreamWriter) WriteHeader(columns []string, widths []int) error {
	widthMap := tw.NewMapper[int, int]()
	for i, w := range widths {
		widthMap[i] = w + 2 // include the renderer's cell padding
	}

	s.table = tablewriter.NewTable(s.out,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Stream: tw.StreamConfig{Enable: true},
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.Off, AutoWrap: tw.WrapTruncate},
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
			},
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapTruncate},
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
			},
			Widths: tw.CellWidth{PerColumn: widthMap},
		}),
	)
	if err := s.table.Start(); err != nil {
		return err
	}

	if s.enableColor {
		colored := make([]string, len(columns))
		for i, c := range columns {
			colored[i] = headerColorPrefix + c + headerColorSuffix
		}
		columns = colored
	}
	s.table.Header(columns)
	return nil
}

func (s *StreamWriter) WriteRow(cells []string) error {
	return s.table.Append(cells)
}

func (s *StreamWriter) Close() error {
	if s.table == nil {
		return nil
	}
	return s.table.Close()
}
