## ADDED Requirements

### Requirement: Default view of an API command
An extension API command SHALL accept an optional `view` property with `columns` (a list of columns using the `--select` syntax) and/or `name` (the name of an existing view). When both are set, `columns` SHALL take precedence. The view SHALL be used to display the command's output when the view option is `auto`.

#### Scenario: Inline columns
- **WHEN** the command sets `view.columns` and the user runs it without `--view`
- **THEN** the output is displayed with those columns, regardless of the views which would match the response

#### Scenario: User selects a view
- **WHEN** the user runs the command with `--view <name>` or `--view off`, or the `defaults.view` setting is not `auto`
- **THEN** the command's view is not used

#### Scenario: User selects properties
- **WHEN** the user runs the command with `--select`
- **THEN** the selected properties are displayed, not the command's view

### Requirement: View name resolution
A view `name` SHALL be resolved in the views of the command's extension first, and then in all views. A name which contains the namespace separator `::` SHALL be resolved as given. When no view matches the name, a warning SHALL be logged and the view SHALL be detected automatically.

#### Scenario: Same view name in two extensions
- **WHEN** the extensions `a` and `b` both have a view named `health`, and a command of extension `b` sets `view.name: health`
- **THEN** the view `b::health` is used

#### Scenario: Unknown view name
- **WHEN** the command sets a view name which does not exist
- **THEN** a warning is logged and the view is detected from the response

### Requirement: Views excluded from automatic detection
A view definition SHALL accept `autoMatch` (default `true`). A view with `"autoMatch": false` SHALL NOT be picked by automatic view detection, and SHALL still be found by name.

#### Scenario: Manual only view
- **WHEN** a view has `"autoMatch": false` and the user runs `c8y api` on a path which the view would match
- **THEN** the view is not used
- **AND** `--view <name>` still uses it
