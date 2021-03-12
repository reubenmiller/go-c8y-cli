package tableviewer

import (
	"io"
	"math"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
)

// TableView renders a table in the terminal
type TableView struct {
	Out            io.Writer
	Columns        []string
	ColumnWidths   []int
	MinColumnWidth int
	MaxColumnWidth int
	ColumnPadding  int
	Data           gjson.Result
	TableData      [][]string
}

func (v *TableView) getValue(value gjson.Result) []string {
	row := []string{}
	lastIndex := len(v.Columns) - 1
	for i, col := range v.Columns {
		columnValue := strings.Trim(value.Get(col).Raw, "\"")

		columnWidth := v.MaxColumnWidth
		if i < len(v.ColumnWidths) {
			columnWidth = v.ColumnWidths[i]
		}
		if columnWidth != 0 && len(columnValue) > columnWidth {
			if i == lastIndex {
				if len(columnValue) > v.MaxColumnWidth {
					columnValue = columnValue[0:v.MaxColumnWidth-3] + "..."
				}
			} else {
				// fmt.Printf("column width: %d, maxColumnWidth: %d, value='%s'\n", len(columnValue), columnWidth, columnValue)
				columnValue = columnValue[0:columnWidth-3] + "..."
			}
		}
		row = append(row, columnValue)

	}
	return row
}

func (v *TableView) calculateColumnWidths(minWidth int, row []string) {
	if len(v.ColumnWidths) == 0 {
		// fmt.Printf("Calculating column widths\n")
		v.ColumnWidths = make([]int, len(row))
		for i, columnValue := range row {
			v.ColumnWidths[i] = int(math.Max(float64(len(columnValue)), float64(minWidth))) + v.ColumnPadding
			// v.ColumnWidths[i] = int(math.Min(float64(v.ColumnWidths[i]), float64(v.MaxColumnWidth)))
		}
	}

}

func (v *TableView) getHeaderRow() []string {
	header := []string{}
	for i, name := range v.Columns {
		width := v.MinColumnWidth
		if i < len(v.ColumnWidths) {
			width = len(name)
			// width = v.ColumnWidths[i]
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
			data = append(data, v.getValue(value))
			return true
		})
	} else if r.IsObject() {
		v.calculateColumnWidths(v.MinColumnWidth, v.getValue(r))
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
	if withHeader {
		table.SetHeader(v.Columns)
		table.Append(v.getHeaderRow())
	}

	for i, width := range v.ColumnWidths {
		table.SetColMinWidth(i, width+v.ColumnPadding)
	}

	table.SetAutoWrapText(true)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()
}
