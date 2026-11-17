## Why

Table column widths are derived purely from the data (the sampled rows, capped by the global `settings.views.columnMinWidth` / `columnMaxWidth`). View authors know better than the heuristic which columns matter: a `description` column can eat the whole terminal, while an `id` or `status` column is always short. There is no way for a view to say so (issue #581).

Sampling up to `settings.views.sampleSize` rows (instead of only the first row) already removed the worst case described in the issue, but view authors still cannot pin, cap, or reserve space for a specific column.

## What Changes

- Each entry of a view's `columns` array accepts either the existing string form **or** an object:
  ```json
  {
    "columns": [
      "id",
      "name",
      { "path": "tags.**", "width": 15 },
      { "path": "description", "maxWidth": 40 },
      { "name": "model", "path": "c8y_Hardware.model", "minWidth": 12 }
    ]
  }
  ```
  - `path` (required): the same selector syntax a string entry uses (wildcards, `!` exclusions).
  - `name` (optional): column alias, equivalent to the existing `name:path` string form.
  - `width` (optional): fixed column width in characters. It wins over the sampled width and the global min/max.
  - `minWidth` / `maxWidth` (optional): per-column bounds that override the global `columnMinWidth` / `columnMaxWidth` for that column; the sampled width is clamped between them.
- String entries keep their exact current meaning. Existing views and `--select` are unaffected (no breaking change).
- Extension API command views (`view.columns` in an extension's API spec YAML) accept the same object form.
- The views JSON schema (`tools/schema/views.json`) documents the object form so editors validate and autocomplete it.

## Capabilities

### New Capabilities
- `view-column-widths`: per-column width rules declared in a view definition, and how they combine with sampled widths, global limits, and the terminal width.

### Modified Capabilities
<!-- None: extension-command-views is still an open change (add-command-views); its requirements are unchanged, it just inherits the richer column type. -->

## Impact

- `pkg/dataview`: `Definition.Columns` changes from `[]string` to a `[]ColumnSpec` type with a custom `UnmarshalJSON` (string or object). `GetView` / `GetViewByName` keep returning the selector strings; a new accessor exposes the width rules of the active view.
- `internal/integration/models/api_spec.go`: extension `view.columns` gets the same type (custom `UnmarshalYAML`).
- `pkg/console`: the table engine receives the active view's width rules.
- **go-c8y** (`github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/encode`): `TableOptions` needs a per-column width hook; `Table.resolveColumns` applies it. This is a cross-repo change and needs a go-c8y release first.
- `tools/schema/views.json`, view docs, and system tests.
