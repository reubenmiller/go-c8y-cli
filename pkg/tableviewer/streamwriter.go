package tableviewer

import (
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
)

// StreamWriter adapts tablewriter's streaming renderer to the go-c8y/v2
// output table engine: the SDK side owns column resolution, width sampling
// and cell fitting (encode.Table) and hands resolved rows to this writer,
// which renders them incrementally — rows are written as they arrive instead
// of buffering the whole result set.
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
	columns     []encode.Column
}

var _ encode.RowWriter = (*StreamWriter)(nil)

func NewStreamWriter(out io.Writer, enableColor bool) *StreamWriter {
	return &StreamWriter{out: out, enableColor: enableColor}
}

// WriteHeader receives the resolved columns (names, sampled widths and
// alignments) from the table engine and starts the underlying streaming
// table. Fixed per-column widths are required for streaming output (rows
// rendered early cannot be re-aligned), which is exactly what the engine's
// sampling provides.
func (s *StreamWriter) WriteHeader(columns []encode.Column) error {
	widthMap := tw.NewMapper[int, int]()
	names := make([]string, len(columns))
	aligns := make([]tw.Align, len(columns))
	for i, col := range columns {
		widthMap[i] = col.Width + 2 // include the renderer's cell padding
		names[i] = col.Name
		aligns[i] = tw.AlignLeft
		if col.Align == encode.AlignRight {
			aligns[i] = tw.AlignRight
		}
	}

	s.columns = columns
	s.table = tablewriter.NewTable(s.out,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithConfig(tablewriter.Config{
			Stream: tw.StreamConfig{Enable: true},
			// Right-aligned cells are padded manually in WriteRow (see
			// below), which trimming would undo.
			Behavior: tw.Behavior{TrimSpace: tw.Off},
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.Off, AutoWrap: tw.WrapTruncate},
				// The markdown separator derives its alignment markers
				// (---:) from the header alignment.
				Alignment: tw.CellAlignment{Global: tw.AlignLeft, PerColumn: aligns},
			},
			Row: tw.CellConfig{
				// Cells are pre-fitted (truncated/wrapped) by the engine;
				// multi-line cells are already split with newlines.
				// Alignment is applied manually in WriteRow: in streaming
				// mode tablewriter only honors per-column alignment for
				// the first appended row.
				Formatting: tw.CellFormatting{AutoWrap: tw.WrapNone},
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
			},
			Widths: tw.CellWidth{PerColumn: widthMap},
		}),
	)
	if err := s.table.Start(); err != nil {
		return err
	}

	if s.enableColor {
		colored := make([]string, len(names))
		for i, c := range names {
			colored[i] = headerColorPrefix + c + headerColorSuffix
		}
		names = colored
	}
	s.table.Header(names)
	return nil
}

func (s *StreamWriter) WriteRow(cells []string) error {
	for i, col := range s.columns {
		if i >= len(cells) || col.Align != encode.AlignRight {
			continue
		}
		// Left-pad right-aligned cells to the column's content width.
		if pad := col.Width - runewidth.StringWidth(cells[i]); pad > 0 {
			cells[i] = strings.Repeat(" ", pad) + cells[i]
		}
	}
	return s.table.Append(cells)
}

func (s *StreamWriter) Close() error {
	if s.table == nil {
		return nil
	}
	return s.table.Close()
}
