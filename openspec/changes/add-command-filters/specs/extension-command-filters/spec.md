## ADDED Requirements

### Requirement: Client side filters of an API command
An extension API command SHALL accept an optional `filter` property, a list of filters using the `--filter` syntax. The filters SHALL be applied to the command's JSON response before it is displayed.

#### Scenario: Command filter
- **WHEN** the command sets `filter: [nothas c8y_Template]` and the response contains items with and without `c8y_Template`
- **THEN** only the items without `c8y_Template` are displayed

#### Scenario: Combined with user filters
- **WHEN** the user runs the command with `--filter "name like *B"`
- **THEN** only the items which match both the command's filters and the user's filters are displayed

#### Scenario: Other commands are not affected
- **WHEN** the user calls the same endpoint with `c8y api`
- **THEN** the command's filters are not applied

