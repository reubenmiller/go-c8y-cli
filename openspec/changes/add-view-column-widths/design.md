## Context

- A view (`pkg/dataview.Definition`) has `Columns []string`. Each string is a selector (pluck pattern): plain paths, wildcards (`tags.**`), exclusions (`!history.**`), and aliases (`name:path`).
- `GetView` / `GetViewByName` return those strings; `cmdutil/factory.go` and `request/request.go` put them into `Filters.Pluck`. After plucking, the table sees flattened column names: the alias if one is set, otherwise the full path (`c8y_Hardware.model`), or the array root for scalar arrays (`tags.**` becomes `tags`).
- The table engine lives in go-c8y (`pkg/c8y/output/encode.Table`). `resolveColumns` computes each width as `max(min(sampled + padding, MaxColumnWidth), MinColumnWidth + padding, len(header))`, then drops trailing columns past `MaxTableWidth`. Options are global; there is no per-column input.
- Extension API commands declare `view.columns` as `[]string` in YAML (`internal/integration/models/api_spec.go`) and are turned into a `dataview.Definition` in `pkg/cmdparser/view.go`.

## Goals / Non-Goals

**Goals:**
- Let a view set a fixed width, or min/max bounds, per column.
- Keep every existing view file, extension spec, and `--select` invocation working unchanged.
- One column type shared by JSON view files and extension YAML.

**Non-Goals:**
- Width syntax on the `--select` flag (for example the issue's `tags.**[width:15]`). Possible later, see Open Questions.
- Per-column alignment, formatting, or colors. The object form leaves room for them, but they are not part of this change.
- Percentage or terminal-relative widths.

## Decisions

### 1. Object entries instead of an inline suffix
Use `{"path": "...", "width": 15}` rather than the issue's `"tags.**[width:15]"`.
- Object form is validated by the JSON schema, so editors catch typos (`widht`) and autocomplete keys.
- No new mini-language to parse, and no risk of clashing with selector characters (`[` already has meaning in gjson-style paths).
- Room to grow (`align`, `format`) without inventing more suffix syntax.
- Alternative considered: a separate `columnWidths: {"description": 40}` map next to `columns`. Rejected because it duplicates the path and drifts out of sync when a column is renamed.

### 2. `ColumnSpec` type with custom unmarshalling
```go
type ColumnSpec struct {
    Name     string `json:"name,omitempty" yaml:"name,omitempty"`
    Path     string `json:"path"           yaml:"path"`
    Width    int    `json:"width,omitempty" yaml:"width,omitempty"`
    MinWidth int    `json:"minWidth,omitempty" yaml:"minWidth,omitempty"`
    MaxWidth int    `json:"maxWidth,omitempty" yaml:"maxWidth,omitempty"`
}
```
- `UnmarshalJSON` / `UnmarshalYAML` accept a bare string (stored as `Path`, with `Name` split off if it has the `name:path` form) or an object.
- `Selector()` returns the string used for plucking today (`name:path` or `path`), so all existing plucking code keeps receiving `[]string`.
- `MarshalJSON` writes a string when no width fields are set, so `c8y settings views` style output and completions look the same as before.
- Lives in `pkg/dataview`; the extension model imports it.

### 3. Keep `GetView` returning `[]string`; expose rules separately
Changing the return type touches every caller and the `Filters.Pluck` plumbing for no benefit. Instead:
- `GetView` / `GetViewByName` return `Definition.Selectors()` (same `[]string` as today).
- The active view is already stored (`DataView.ActiveView`, set by `GetView` and `SetActiveView`). Add `Columns.WidthRule(name)` returning the first entry with a width that matches a rendered column.
- For a manual `--view name`, the definition is looked up again with `GetDefinitionByName` so it gets widths too.

### 4. Matching rules to rendered column names
The table only knows rendered names (`c8y_Hardware.model`, `model`, `tags`), not the selector that produced them. A rule matches a rendered column when:
1. the rule has `name` and it equals the column name (or is the prefix of a column from a wildcard path), or
2. the rule's `path` equals the column name, or
3. the rule's `path` contains a wildcard and either wildcard-matches the column name (`pkg/matcher`), or the column name equals the path with a trailing `.**` / `.*` removed (covers scalar arrays collapsed to their root, such as `tags.**` to `tags`).

First matching rule in `columns` order wins. Exclusion entries (`!…`) never carry rules.
- Alternative considered: record the selector-to-column mapping inside the pluck step. More exact, but it threads new state through `jsonfilter` and every output path. Revisit only if the matching heuristic proves wrong in practice.

### 5. Engine hook in go-c8y: a per-column callback
Add to `encode.TableOptions`:
```go
// ColumnWidth returns per-column width constraints for a resolved column name.
// Zero values mean "use the global option".
ColumnWidth func(name string) ColumnWidthRule

type ColumnWidthRule struct{ Width, MinWidth, MaxWidth int }
```
All three values are the final column width (column padding is not added), and a per-column `MaxWidth` also wins over the global minimum.
In `resolveColumns`, per column:
- `minW`, `maxW` default to `MinColumnWidth` / `MaxColumnWidth`, overridden by non-zero hook values.
- If `width > 0`: `colWidth = max(width + padding, len(name))`.
- Else: today's formula with `minW` / `maxW`.
- Pass the per-column width to `Transform` (already happens because the transform receives `Column.Width`).

A callback (rather than a `map[string]...`) keeps the engine unaware of wildcard matching and works with lazily resolved columns. go-c8y-cli supplies a closure built from the active view's `WidthRules()` and the matching in Decision 4.

### 6. Where the console gets the rules
`Console` already has `tableOptions` and builds the engine lazily in `tableEngine()`. Add `Console.SetColumnWidthRules(func(name string) (int, int, int))`, called from `cmdutil/factory.go` at the point where the view is resolved and `Filters.Pluck` is set. When `--select` is set or views are off, the rules are cleared (the spec requires `--select` to bypass view widths).

## Risks / Trade-offs

- [Cross-repo change: the engine is in go-c8y] → Land the `ColumnWidth` hook in go-c8y first as an additive, zero-value-safe option, then bump the dependency here.
- [Heuristic matching (Decision 4) may miss an edge case, for example a wildcard entry whose flattened names differ from the pattern] → Covered by unit tests on the matcher with the shapes used in `examples/views/default`; a miss only means the column falls back to today's width, never an error.
- [A fixed width wider than the terminal] → Existing `MaxTableWidth` logic still drops or shrinks trailing columns, so output never overflows.
- [Third-party tools reading view files as `[]string`] → Only views that opt into the object form are affected; the default views shipped with the CLI can keep using strings except where a width is useful.

## Migration Plan

No migration. String entries are unchanged. View authors opt in per column. Rollback is reverting the change; views using object entries would then fail to load and fall back to auto-detection, so default views should only adopt the object form in the same release.

## Open Questions

- Should `--select` also gain inline width syntax (for example `description:40` is taken by aliases, so something like `description[w=40]`)? Proposed: no for now, revisit if asked.
- Should a fixed `width` also be allowed to exceed `columnMaxWidth`? This proposal says yes (explicit beats global). Confirm.
- Which default views should adopt widths in this change (candidates: `description`, `text`, `name` columns in alarms, events, and operations views)?
