## ADDED Requirements

### Requirement: Column entries accept a string or an object
A view's `columns` array SHALL accept, per entry, either a string (current selector syntax, including the `name:path` alias form) or an object with a required `path` and optional `name`, `width`, `minWidth`, and `maxWidth`. A string entry SHALL behave exactly as it does today.

#### Scenario: Mixed string and object entries
- **WHEN** a view defines `"columns": ["id", {"path": "name", "width": 20}]`
- **THEN** the view loads without error and the table shows the `id` and `name` columns in that order

#### Scenario: Object with alias
- **WHEN** a view defines `{"name": "model", "path": "c8y_Hardware.model"}`
- **THEN** the column is selected and labelled exactly as with the string `"model:c8y_Hardware.model"`

#### Scenario: Invalid object entry
- **WHEN** a column object has no `path`, or a width value that is not a positive integer, or `minWidth` greater than `maxWidth`
- **THEN** a warning is logged and the definitions of that view file are not loaded, the same as other malformed view files

### Requirement: Fixed column width
When a column entry sets `width`, the rendered column SHALL be exactly that many characters wide (column padding is not added), regardless of the sampled values and the global `columnMinWidth` / `columnMaxWidth` settings. Values longer than the width SHALL be truncated or wrapped according to the current row mode.

#### Scenario: Long values are truncated to the fixed width
- **WHEN** a view sets `{"path": "description", "width": 10}` and a row's description is 50 characters, with row mode `truncate`
- **THEN** the description column is 10 characters wide and the value is truncated with the usual ellipsis marker

#### Scenario: Short values keep the fixed width
- **WHEN** a view sets `{"path": "id", "width": 12}` and all sampled ids are 4 characters long
- **THEN** the id column is still 12 characters wide

### Requirement: Per-column width bounds
When a column entry sets `minWidth` and/or `maxWidth` (without `width`), the column width SHALL be the sampled width clamped to those bounds. A bound that is not set SHALL fall back to the corresponding global setting. If both `width` and a bound are set, `width` SHALL win.

#### Scenario: Per-column max overrides global max
- **WHEN** `settings.views.columnMaxWidth` is 80, a view sets `{"path": "description", "maxWidth": 30}`, and sampled descriptions are 70 characters
- **THEN** the description column is 30 characters wide

#### Scenario: Per-column min reserves space
- **WHEN** a view sets `{"path": "status", "minWidth": 12}` and all sampled statuses are 2 characters long
- **THEN** the status column is 12 characters wide

### Requirement: Width rules apply to every column a pattern expands to
A width rule on an entry whose `path` contains wildcards SHALL apply to every table column produced by that entry. A rule on an aliased entry SHALL also apply when the column is displayed under its path (which happens when the value does not exist). If a table column is produced by more than one entry, the rule of the first matching entry in the `columns` array SHALL apply.

#### Scenario: Wildcard entry
- **WHEN** a view sets `{"path": "c8y_Position.*", "width": 10}` and the data contains `c8y_Position.lat` and `c8y_Position.lng`
- **THEN** both resulting columns are 10 characters wide

### Requirement: Header and terminal width still apply
A column SHALL never be narrower than its header text. The total table width limit (terminal width) SHALL still drop trailing columns that do not fit, the same as today.

#### Scenario: Width smaller than the header
- **WHEN** a view sets `{"path": "lastUpdated", "width": 5}`
- **THEN** the column is as wide as the header `lastUpdated` (11 characters)

### Requirement: Non-table output is unaffected
Width rules SHALL only affect table output. JSON, CSV, TSV, and other output formats SHALL produce the same output as an equivalent view whose entries are plain strings.

#### Scenario: JSON output ignores widths
- **WHEN** a view with `{"path": "description", "width": 10}` is used with `--output json`
- **THEN** the description value is output in full

### Requirement: Extension command views accept the object form
The `view.columns` list of an extension API command SHALL accept the same string-or-object entries as a view definition file, with the same width behavior.

#### Scenario: Extension command with a fixed-width column
- **WHEN** an extension API command declares `view: {columns: [id, {path: description, width: 20}]}` and the command is run with table output
- **THEN** the description column is 20 characters wide

### Requirement: Explicit selection bypasses view widths
When the user passes `--select` (or the view is turned off), the view's width rules SHALL NOT be applied, since the columns no longer come from the view.

#### Scenario: --select overrides the view
- **WHEN** a view sets `{"path": "name", "width": 5}` and the user runs the command with `--select id,name`
- **THEN** the name column uses the sampled width
