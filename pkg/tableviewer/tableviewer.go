package tableviewer

import (
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/ts"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/gjsonpath"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/numbers"
	"github.com/tidwall/gjson"
)

var Logger *log.Logger

var TextEllipsis = "…"

func init() {
	Logger = log.New(io.Discard, "tableviewer", 0)
}

const (
	RowModeTruncate = "truncate"
	RowModeWrap     = "wrap"
	RowModeOverflow = "overflow" // default
)

// TableView renders a table in the terminal
type TableView struct {
	Out                      io.Writer
	Columns                  []string
	ColumnWidths             []int
	ColumnAlignments         []int
	MinColumnWidth           int
	MinEmptyValueColumnWidth int
	MaxColumnWidth           int
	ColumnPadding            int
	SampleSize               int
	Data                     gjson.Result
	TableData                [][]string
	EnableColor              bool
	RowMode                  string
	NumberFormatter          numbers.NumberFormatter
}

// DefaultSampleSize maximum number of rows sampled when resolving column widths
const DefaultSampleSize = 5

func (v *TableView) sampleLimit() int {
	if v.SampleSize > 0 {
		return v.SampleSize
	}
	return DefaultSampleSize
}

func (v *TableView) getValue(value gjson.Result) []string {
	row, alignments := v.getRow(value)
	if v.ColumnAlignments == nil {
		v.ColumnAlignments = alignments
	}
	return row
}

func (v *TableView) getRow(value gjson.Result) (row []string, alignments []int) {
	row = []string{}
	alignments = []int{}

	for i, col := range v.Columns {
		node := value.Get(gjsonpath.EscapePath(col))

		columnValue := ""
		columnAlignment := tablewriter.ALIGN_LEFT
		if node.Type == gjson.Number {
			columnAlignment = tablewriter.ALIGN_RIGHT
			columnValue = v.NumberFormatter.Display(node.Float(), node.Raw, "")
		} else {
			columnValue = strings.Trim(node.Raw, "\"")
		}

		alignments = append(alignments, columnAlignment)

		columnWidth := v.MaxColumnWidth
		if i < len(v.ColumnWidths) {
			columnWidth = v.ColumnWidths[i]
		}
		if columnWidth != 0 && len(columnValue) > columnWidth {
			if v.RowMode == RowModeTruncate {
				columnValue = columnValue[0:columnWidth-1] + TextEllipsis
			} else if v.RowMode == RowModeWrap {
				columnValue = WrapLine(columnValue, columnWidth, "")
			}
		}
		row = append(row, columnValue)

	}
	return row, alignments
}

func (v *TableView) getWidth(defaultWidth int) int {
	termSize, err := ts.GetSize()
	if err != nil {
		return defaultWidth
	}
	return termSize.Col() - 1
}

var TABLE_MAX_WIDTH = 120

func minmax(values []int) (min int, max int) {
	for _, val := range values {
		if val > max {
			max = val
		}
		if val < min {
			min = val
		}
	}
	return
}

// cellDisplayWidth display width of a cell value, using the widest line
// for multi-line (e.g. wrapped) values
func cellDisplayWidth(s string) int {
	if !strings.Contains(s, "\n") {
		return tablewriter.DisplayWidth(s)
	}
	width := 0
	for _, line := range strings.Split(s, "\n") {
		if w := tablewriter.DisplayWidth(line); w > width {
			width = w
		}
	}
	return width
}

func (v *TableView) calculateColumnWidths(minWidth int, rows [][]string) {
	if len(v.ColumnWidths) == 0 {
		maxTableWidth := v.getWidth(TABLE_MAX_WIDTH)
		v.ColumnWidths = make([]int, 0)
		columnSeparatorWidth := 3
		tableEndBuffer := 3
		usedWith := 0
		curMinWidth := 0
		columns := []string{}

		// only include columns if they fit in the view
		for i := range v.Columns {
			// use the widest value of the sampled rows, as the first row
			// may not contain a value for every column
			cellWidth := 0
			hasValue := false
			for _, row := range rows {
				if i >= len(row) {
					continue
				}
				if row[i] != "" {
					hasValue = true
				}
				if w := cellDisplayWidth(row[i]); w > cellWidth {
					cellWidth = w
				}
			}

			curMinWidth = minWidth
			if !hasValue && v.MinEmptyValueColumnWidth > 0 {
				curMinWidth = v.MinEmptyValueColumnWidth
			}

			// only pad value (not column widths)
			paddedCellWidth := cellWidth + v.ColumnPadding
			if v.MaxColumnWidth > 0 && paddedCellWidth > v.MaxColumnWidth {
				paddedCellWidth = v.MaxColumnWidth
			}

			Logger.Printf("iColumn: name=%s, cellWidth=%d, min=%d, col=%d",
				v.Columns[i],
				paddedCellWidth,
				curMinWidth+v.ColumnPadding,
				tablewriter.DisplayWidth(v.Columns[i]),
			)

			_, colWidth := minmax([]int{
				paddedCellWidth,
				curMinWidth + v.ColumnPadding,
				tablewriter.DisplayWidth(v.Columns[i]),
			})

			// TODO: When the column value is empty, then use a dedicate empty width value instead
			// of the minimum width value
			if usedWith+colWidth+columnSeparatorWidth > maxTableWidth {
				leftOver := maxTableWidth - usedWith - columnSeparatorWidth - tableEndBuffer
				Logger.Printf("Left over: %d", leftOver)
				if leftOver > curMinWidth {
					v.ColumnWidths = append(v.ColumnWidths, leftOver)
					columns = append(columns, v.Columns[i])
				}
				break
			}
			v.ColumnWidths = append(v.ColumnWidths, colWidth)
			usedWith += colWidth + columnSeparatorWidth // overhead
			if i < len(v.ColumnWidths) {
				columns = append(columns, v.Columns[i])
			}
		}
		v.Columns = columns

		Logger.Printf("columns: %v\n", v.Columns)
		Logger.Printf("column widths: %v\n", v.ColumnWidths)
		Logger.Printf("Column summary: max=%d, used=%d", maxTableWidth, usedWith)
	}
}

// SampleColumnWidths resolve the column widths and alignments from multiple
// sampled rows, where each row is a json object or an array of objects.
// It does nothing if the column widths have already been resolved
func (v *TableView) SampleColumnWidths(jsonRows [][]byte) {
	if len(v.ColumnWidths) != 0 {
		return
	}
	samples := []gjson.Result{}
	limit := v.sampleLimit()
	for _, b := range jsonRows {
		r := gjson.ParseBytes(b)
		if r.IsArray() {
			for _, item := range r.Array() {
				if len(samples) >= limit {
					break
				}
				samples = append(samples, item)
			}
		} else if r.IsObject() {
			samples = append(samples, r)
		}
		if len(samples) >= limit {
			break
		}
	}
	v.primeFromSamples(samples)
}

// primeFromSamples resolve the column widths and alignments from sampled rows
func (v *TableView) primeFromSamples(samples []gjson.Result) {
	if len(samples) == 0 || (len(v.ColumnWidths) != 0 && v.ColumnAlignments != nil) {
		return
	}
	rows := make([][]string, 0, len(samples))
	alignments := make([][]int, 0, len(samples))
	for _, sample := range samples {
		row, rowAlignments := v.getRow(sample)
		rows = append(rows, row)
		alignments = append(alignments, rowAlignments)
	}

	if v.ColumnAlignments == nil {
		// use the alignment of the first row which has a value for the column,
		// so columns are still aligned correctly (e.g. numbers) even if the
		// first row does not contain a value
		merged := make([]int, len(v.Columns))
		for i := range v.Columns {
			merged[i] = tablewriter.ALIGN_LEFT
			for ri, row := range rows {
				if i < len(row) && row[i] != "" && i < len(alignments[ri]) {
					merged[i] = alignments[ri][i]
					break
				}
			}
		}
		v.ColumnAlignments = merged
	}

	v.calculateColumnWidths(v.MinColumnWidth, rows)
}

func (v *TableView) getHeaderRow() []string {
	header := []string{}
	for i, name := range v.Columns {
		width := v.MinColumnWidth
		if i < len(v.ColumnWidths) {
			width = len(name)
		}
		header = append(header, strings.Repeat("-", width))
	}
	return header
}

// TransformData transform the data so that is presentable in the terminal
func (v *TableView) TransformData(j []byte, property string) [][]string {
	r := gjson.ParseBytes(j)
	data := [][]string{}
	if property != "" {
		if v := r.Get(property); v.Exists() {
			r = v
		}
	}

	if r.IsArray() {
		if items := r.Array(); len(items) > 0 {
			sampleCount := min(len(items), v.sampleLimit())
			v.primeFromSamples(items[0:sampleCount])
		}
		r.ForEach(func(key, value gjson.Result) bool {
			Logger.Printf("parsing row: columns: %v", v.Columns)
			data = append(data, v.getValue(value))
			return true
		})
	} else if r.IsObject() {
		v.primeFromSamples([]gjson.Result{r})
		Logger.Printf("parsing row: columns: %v", v.Columns)
		data = append(data, v.getValue(r))
	}

	return data
}

func WrapLine(line string, width int, wrapPrefix string) string {
	if len(line) <= width {
		return line
	}
	name := strings.Builder{}
	lineWidth := 0
	for _, c := range line {
		if unicode.IsSpace(c) {
			// table writer will split on whitespace
			lineWidth = -1
		}
		if lineWidth >= width {
			name.WriteRune('\n')
			if wrapPrefix != "" {
				name.WriteString(wrapPrefix)
				lineWidth = 1
			} else {
				lineWidth = 0
			}

		}
		name.WriteRune(c)
		lineWidth++
	}
	return name.String()
}

func (v *TableView) GetHeaders() (headers []string) {
	for i, col := range v.Columns {
		if i < len(v.ColumnWidths) {
			headers = append(headers, WrapLine(col, v.ColumnWidths[i], ""))
		}
	}
	return headers
}

// Render writes the json data to console in the form of a table
func (v *TableView) Render(jsonData []byte, withHeader bool) {
	data := v.TransformData(jsonData, "")

	if v.Out == nil {
		v.Out = os.Stdout
	}
	table := tablewriter.NewWriter(v.Out)

	isMarkdown := true
	if withHeader {
		table.SetHeader(v.GetHeaders())
		if !isMarkdown {
			table.Append(v.getHeaderRow())
		}
	}

	maxWidth := 0
	headerColors := []tablewriter.Colors{}
	for i, width := range v.ColumnWidths {
		headerColors = append(headerColors, tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor})
		table.SetColMinWidth(i, width)
		if width >= maxWidth {
			maxWidth = width
		}
	}
	table.SetColWidth(maxWidth)

	if withHeader && v.EnableColor {
		table.SetHeaderColor(headerColors...)
	}

	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment(v.ColumnAlignments)

	wrapEnabled := v.RowMode == RowModeWrap

	if isMarkdown {
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.SetAutoFormatHeaders(false)
		table.SetAutoWrapText(true)
	} else {
		table.SetAutoWrapText(wrapEnabled)
		table.SetReflowDuringAutoWrap(wrapEnabled)
		table.SetAutoFormatHeaders(false)

		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowLine(false)
		table.SetRowSeparator("-")
		table.SetTablePadding(" ")
		table.SetNoWhiteSpace(true)
	}

	// Enable row separator when wrapping cells to make it easier to read
	table.SetRowSeparator("-")
	table.SetAutoWrapText(wrapEnabled)
	table.SetRowLine(wrapEnabled)

	table.AppendBulk(data)
	table.Render()
}
