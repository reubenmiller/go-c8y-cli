package tableviewer

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode"
)

// ColumnWidthRules returns the per-column width rules of a view definition
// for the table engine. nil is returned if the view does not define any widths
func ColumnWidthRules(definition *dataview.Definition) func(name string) encode.ColumnWidthRule {
	if definition == nil || !definition.Columns.HasWidthRules() {
		return nil
	}
	columns := definition.Columns
	return func(name string) encode.ColumnWidthRule {
		col, ok := columns.WidthRule(name)
		if !ok {
			return encode.ColumnWidthRule{}
		}
		return encode.ColumnWidthRule{
			Width:    col.Width,
			MinWidth: col.MinWidth,
			MaxWidth: col.MaxWidth,
		}
	}
}
