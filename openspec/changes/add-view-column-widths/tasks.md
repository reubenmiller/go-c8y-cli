## 1. Table engine (go-c8y)

- [x] 1.1 Add `ColumnWidth func(name string) ColumnWidthRule` to `encode.TableOptions`
- [x] 1.2 Apply per-column width, min, and max in `Table.resolveColumns` (header width and `MaxTableWidth` still apply)
- [x] 1.3 Unit tests: fixed width, per-column min/max, nil hook keeps current behavior
- [x] 1.4 Merge go-c8y (reubenmiller/go-c8y#179) and bump the dependency in go-c8y-cli

## 2. Column spec type

- [x] 2.1 Add `dataview.ColumnSpec` with `UnmarshalJSON`, `UnmarshalYAML`, `MarshalJSON`, and `Selector()`
- [x] 2.2 Validate entries (path required, positive widths, `minWidth <= maxWidth`) and skip invalid view files with a warning
- [x] 2.3 Change `Definition.Columns` to `[]ColumnSpec`; add `Columns.Selectors()` and `Columns.WidthRule(name)`
- [x] 2.4 Update `GetView`, `GetViewByName`, and view completion (`pkg/cmdutil/viewcompletion.go`) to use `Selectors()`
- [x] 2.5 Change extension `view.columns` in `internal/integration/models/api_spec.go` to `[]dataview.ColumnSpec` and update `pkg/cmdparser/view.go`
- [x] 2.6 Unit tests: string entries, alias strings, object entries, mixed arrays, invalid objects, YAML and JSON

## 3. Wiring widths into table output

- [x] 3.1 Implement rule-to-column matching (exact name, exact path, wildcard, collapsed array root; first match wins) with unit tests
- [x] 3.2 Add `Console.SetColumnWidthRules` and pass it to the engine's `ColumnWidth` option
- [x] 3.3 Set the rules where the view is resolved in `pkg/cmdutil/factory.go` (auto and named view); they are not set for `--select` and `--view off`
- [x] 3.4 Check the request path in `pkg/request/request.go` uses the same wiring

## 4. Schema, docs, and default views

- [x] 4.1 Update `tools/schema/views.json`: `columns.items` is `oneOf` string or column object
- [x] 4.2 Document the object form in the views docs with an example
- [ ] 4.3 Optionally adopt widths in a few default views (long `description` or `text` columns), left as a follow-up

## 5. System tests

- [x] 5.1 Add a test view with fixed, min, and max widths and assert table output column widths
- [x] 5.2 Assert `--output json` and `--select` ignore the width rules
- [x] 5.3 Assert an extension API command with an object column keeps the fixed width (unit test in `pkg/cmdparser`; the system tests use a view from the kitchensink extension)
