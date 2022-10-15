package tableviewer

import (
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/ts"
	"github.com/tidwall/gjson"
)

var Logger *log.Logger

func init() {
	Logger = log.New(io.Discard, "tableviewer", 0)
}

// TableView renders a table in the terminal
type TableView struct {
	Out                      io.Writer
	Columns                  []string
	ColumnWidths             []int
	MinColumnWidth           int
	MinEmptyValueColumnWidth int
	MaxColumnWidth           int
	ColumnPadding            int
	Data                     gjson.Result
	TableData                [][]string
	EnableColor              bool
	EnableTruncate           bool
	EnableWrap               bool
}

func (v *TableView) getValue(value gjson.Result) []string {
	row := []string{}
	for i, col := range v.Columns {
		columnValue := strings.Trim(value.Get(col).Raw, "\"")

		columnWidth := v.MaxColumnWidth
		if i < len(v.ColumnWidths) {
			columnWidth = v.ColumnWidths[i]
		}
		if columnWidth != 0 && len(columnValue) > columnWidth {
			if v.EnableTruncate {
				columnValue = columnValue[0:columnWidth-1] + "…"
			} else if v.EnableWrap {
				// Force wrapping column values by splitting on length (not just whitespace)
				splitValue := strings.Builder{}
				lineWidth := 0
				for _, c := range columnValue {
					if unicode.IsSpace(c) {
						// table writer will split on whitespace
						lineWidth = -1
					}

					if lineWidth >= columnWidth {
						splitValue.WriteString("\n…")
						lineWidth = 1
					}

					splitValue.WriteRune(c)
					lineWidth += 1
				}

				columnValue = splitValue.String()
			}
		}
		row = append(row, columnValue)

	}
	return row
}

func (v *TableView) getWidth(defaultWidth int) int {
	termSize, err := ts.GetSize()
	if err != nil {
		return defaultWidth
	}
	return termSize.Col()
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

func (v *TableView) calculateColumnWidths(minWidth int, row []string) {
	if len(v.ColumnWidths) == 0 {
		maxTableWidth := v.getWidth(TABLE_MAX_WIDTH)
		v.ColumnWidths = make([]int, 0)
		columnSeperatorWidth := 3
		tableEndBuffer := 3
		usedWith := 0
		curMinWidth := 0
		columns := []string{}

		// only include columns if they fit in the view
		for i, columnValue := range row {
			curMinWidth = minWidth
			if columnValue == "" && v.MinEmptyValueColumnWidth > 0 {
				curMinWidth = v.MinEmptyValueColumnWidth
			}

			_, colWidth := minmax([]int{
				// only pad value (not column widths)
				tablewriter.DisplayWidth(columnValue) + v.ColumnPadding,
				curMinWidth + v.ColumnPadding,
				tablewriter.DisplayWidth(v.Columns[i]),
			})

			// TODO: When the column value is empty, then use a dedicate empty width value instead
			// of the minimum width value
			if usedWith+colWidth+columnSeperatorWidth > maxTableWidth {
				leftOver := maxTableWidth - usedWith - columnSeperatorWidth - tableEndBuffer
				Logger.Printf("Left over: %d", leftOver)
				if leftOver > curMinWidth {
					v.ColumnWidths = append(v.ColumnWidths, leftOver)
					columns = append(columns, v.Columns[i])
				}
				break
			}
			v.ColumnWidths = append(v.ColumnWidths, colWidth)
			usedWith += colWidth + 3 // overhead
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

// TransformData tranform the data so that is presentable in the terminal
func (v *TableView) TransformData(j []byte, property string) [][]string {
	r := gjson.ParseBytes(j)
	data := [][]string{}
	if property != "" {
		if v := r.Get(property); v.Exists() {
			r = v
		}
	}

	if r.IsArray() {
		if len(r.Array()) > 0 {
			v.calculateColumnWidths(v.MinColumnWidth, v.getValue(r.Array()[0]))
		}
		r.ForEach(func(key, value gjson.Result) bool {
			Logger.Printf("parsing row: columns: %v", v.Columns)
			data = append(data, v.getValue(value))
			return true
		})
	} else if r.IsObject() {
		v.calculateColumnWidths(v.MinColumnWidth, v.getValue(r))
		Logger.Printf("parsing row: columns: %v", v.Columns)
		data = append(data, v.getValue(r))
	}

	return data
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
		table.SetHeader(v.Columns)
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

	if isMarkdown {
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.SetAutoFormatHeaders(false)
		table.SetAutoWrapText(true)
	} else {
		table.SetAutoWrapText(v.EnableWrap)
		table.SetReflowDuringAutoWrap(v.EnableWrap)
		table.SetAutoFormatHeaders(false)

		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetTablePadding(" ")
		table.SetNoWhiteSpace(true)
	}
	table.AppendBulk(data)
	table.Render()
}
