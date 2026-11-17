package dataview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"gopkg.in/yaml.v3"
)

// ColumnSpec is a single entry of a view's columns. It is either a selector
// string (the same syntax as the --select flag, e.g. "id", "tags.**" or the
// alias form "model:c8y_Hardware.model"), or an object which can also
// constrain the column width in table output:
//
//	{"name": "model", "path": "c8y_Hardware.model", "width": 20}
type ColumnSpec struct {
	// Name is an optional column alias
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Path is the selector of the property (or properties) to display
	Path string `json:"path" yaml:"path"`

	// Width is a fixed column width in characters
	Width int `json:"width,omitempty" yaml:"width,omitempty"`

	// MinWidth is the minimum column width in characters
	MinWidth int `json:"minWidth,omitempty" yaml:"minWidth,omitempty"`

	// MaxWidth is the maximum column width in characters
	MaxWidth int `json:"maxWidth,omitempty" yaml:"maxWidth,omitempty"`
}

// NewColumnSpecs creates column specs from selector strings
func NewColumnSpecs(selectors ...string) []ColumnSpec {
	columns := make([]ColumnSpec, 0, len(selectors))
	for _, selector := range selectors {
		columns = append(columns, ColumnSpec{Path: selector})
	}
	return columns
}

// HasWidth checks if the column constrains the column width
func (c ColumnSpec) HasWidth() bool {
	return c.Width > 0 || c.MinWidth > 0 || c.MaxWidth > 0
}

// Selector returns the selector string used to select the column's properties
func (c ColumnSpec) Selector() string {
	if c.Name != "" {
		return c.Name + ":" + c.Path
	}
	return c.Path
}

// String returns the selector string
func (c ColumnSpec) String() string {
	return c.Selector()
}

// Validate checks the column for invalid values
func (c ColumnSpec) Validate() error {
	if strings.TrimSpace(c.Path) == "" {
		return fmt.Errorf("column path must not be empty")
	}
	if c.Width < 0 || c.MinWidth < 0 || c.MaxWidth < 0 {
		return fmt.Errorf("column widths must be positive. path=%s", c.Path)
	}
	if c.MinWidth > 0 && c.MaxWidth > 0 && c.MinWidth > c.MaxWidth {
		return fmt.Errorf("column minWidth must not be greater than maxWidth. path=%s, minWidth=%d, maxWidth=%d", c.Path, c.MinWidth, c.MaxWidth)
	}
	if c.HasWidth() && (strings.HasPrefix(c.Path, "!") || strings.Contains(c.Path, ",")) {
		return fmt.Errorf("column widths can not be used with exclusions or multiple selectors. path=%s", c.Path)
	}
	return nil
}

// Matches checks if a rendered table column was produced by the column spec.
// The column name is either the alias, the property path, or (for wildcard
// paths) any property path matching the wildcard, including the root of a
// wildcard, e.g. "tags" for "tags.**" (scalar arrays are displayed under the root)
func (c ColumnSpec) Matches(column string) bool {
	if c.Name != "" {
		if strings.EqualFold(c.Name, column) {
			return true
		}
		// an alias of a wildcard path is used as a prefix
		if strings.ContainsAny(c.Path, "*") && hasPathPrefix(column, c.Name) {
			return true
		}
		// the path is displayed instead of the alias when the value does not exist
	}
	if strings.EqualFold(c.Path, column) {
		return true
	}
	if !strings.ContainsAny(c.Path, "*") {
		return false
	}
	if match, err := matcher.MatchWithWildcards(column, c.Path); err == nil && match {
		return true
	}
	root := strings.TrimSuffix(strings.TrimSuffix(c.Path, ".**"), ".*")
	return root != c.Path && strings.EqualFold(root, column)
}

func hasPathPrefix(column, prefix string) bool {
	return len(column) > len(prefix) &&
		strings.EqualFold(column[:len(prefix)], prefix) &&
		column[len(prefix)] == '.'
}

// UnmarshalJSON accepts either a selector string or a column object
func (c *ColumnSpec) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '"' {
		var selector string
		if err := json.Unmarshal(data, &selector); err != nil {
			return err
		}
		*c = ColumnSpec{Path: selector}
		return nil
	}

	type columnSpec ColumnSpec
	var value columnSpec
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("invalid view column. expected a string or an object with path, name, width, minWidth and maxWidth. %w", err)
	}
	*c = ColumnSpec(value)
	return c.Validate()
}

// MarshalJSON writes the column as a string if it does not use any object properties
func (c ColumnSpec) MarshalJSON() ([]byte, error) {
	if c.Name == "" && !c.HasWidth() {
		return json.Marshal(c.Path)
	}
	type columnSpec ColumnSpec
	return json.Marshal(columnSpec(c))
}

// UnmarshalYAML accepts either a selector string or a column object
func (c *ColumnSpec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		*c = ColumnSpec{Path: node.Value}
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("invalid view column (line %d). expected a string or an object", node.Line)
	}
	allowed := map[string]bool{"name": true, "path": true, "width": true, "minWidth": true, "maxWidth": true}
	for i := 0; i < len(node.Content); i += 2 {
		if key := node.Content[i].Value; !allowed[key] {
			return fmt.Errorf("invalid view column (line %d). unknown property %q", node.Line, key)
		}
	}
	type columnSpec ColumnSpec
	var value columnSpec
	if err := node.Decode(&value); err != nil {
		return err
	}
	*c = ColumnSpec(value)
	return c.Validate()
}

// MarshalYAML writes the column as a string if it does not use any object properties
func (c ColumnSpec) MarshalYAML() (any, error) {
	if c.Name == "" && !c.HasWidth() {
		return c.Path, nil
	}
	type columnSpec ColumnSpec
	return columnSpec(c), nil
}

// Columns is a list of view columns
type Columns []ColumnSpec

// Selectors returns the selector strings of the columns
func (c Columns) Selectors() []string {
	selectors := make([]string, 0, len(c))
	for _, col := range c {
		selectors = append(selectors, col.Selector())
	}
	return selectors
}

// WidthRule returns the first column with width constraints which matches the
// rendered column name
func (c Columns) WidthRule(column string) (ColumnSpec, bool) {
	for _, col := range c {
		if col.HasWidth() && col.Matches(column) {
			return col, true
		}
	}
	return ColumnSpec{}, false
}

// HasWidthRules checks if any column constrains its width
func (c Columns) HasWidthRules() bool {
	for _, col := range c {
		if col.HasWidth() {
			return true
		}
	}
	return false
}
