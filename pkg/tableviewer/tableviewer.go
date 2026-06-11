// Package tableviewer renders terminal tables. The column resolution, width
// sampling and row streaming is handled by the go-c8y/v2 output table engine
// (encode.Table); this package contributes the go-c8y-cli specific pieces:
// cell formatting (number formatting, escaped paths), row fitting modes
// (truncate/wrap/overflow) and the tablewriter-based streaming renderer
// (StreamWriter).
package tableviewer

import (
	"strings"
	"unicode"

	"github.com/olekukonko/ts"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/gjsonpath"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/numbers"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
	"github.com/tidwall/gjson"
)

var TextEllipsis = "…"

const (
	RowModeTruncate = "truncate"
	RowModeWrap     = "wrap"
	RowModeOverflow = "overflow" // default
)

const (
	headerColorPrefix = "\x1b[1m\x1b[36m" // bold cyan
	headerColorSuffix = "\x1b[0m"
)

// TerminalWidth returns the usable width of the terminal, or defaultWidth
// when it can not be detected (e.g. output is not a terminal)
func TerminalWidth(defaultWidth int) int {
	termSize, err := ts.GetSize()
	if err != nil {
		return defaultWidth
	}
	return termSize.Col() - 1
}

// CellFormatter returns the cell extraction used for go-c8y-cli tables:
// column paths may contain special characters (escaped for gjson lookups),
// numbers are formatted with the given formatter and aligned right, and
// other values are rendered from their raw JSON representation
func CellFormatter(formatter numbers.NumberFormatter) encode.CellFormatter {
	return func(doc jsondoc.JSONDoc, column string) (string, encode.Align) {
		node := doc.Get(gjsonpath.EscapePath(column))
		if node.Type == gjson.Number {
			value := node.String()
			if formatter != nil {
				value = formatter.Display(node.Float(), node.Raw, "")
			}
			return value, encode.AlignRight
		}
		return strings.Trim(node.Raw, "\""), encode.AlignLeft
	}
}

// CellTransform returns the cell fitting behavior for the given row mode:
// truncate (ellipsis marker), wrap (multi-line cells) or overflow (values
// are rendered untouched, the default)
func CellTransform(rowMode string) encode.CellTransform {
	switch rowMode {
	case RowModeTruncate:
		return func(cell string, width int) string {
			if width != 0 && len(cell) > width {
				return cell[0:width-1] + TextEllipsis
			}
			return cell
		}
	case RowModeWrap:
		return func(cell string, width int) string {
			if width != 0 && len(cell) > width {
				return WrapLine(cell, width, "")
			}
			return cell
		}
	default:
		return nil
	}
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
