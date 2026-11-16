# extension-json-output Specification

## Purpose
TBD - created by archiving change add-binary-extensions. Update Purpose after archive.
## Requirements
### Requirement: JSON output commands use the c8y output pipeline
For a declared binary command with `output: json`, the system SHALL capture the binary's stdout, parse it as a JSON array, a single JSON object, or newline-delimited JSON objects, and write each item through the same output handling as built-in commands (views, `--select`, `--filter`, `--output`, `--compact`, `--raw`).

#### Scenario: Table output
- **WHEN** the binary writes two JSON objects as NDJSON and the user runs `c8y foo services list --output table`
- **THEN** c8y prints a table with two rows using the matching view

#### Scenario: Select
- **WHEN** the user runs `c8y foo services list --select id,name`
- **THEN** each item only contains `id` and `name`

#### Scenario: Extension view
- **WHEN** the extension ships `views/services.json` that matches the items
- **THEN** that view selects the default table columns

### Requirement: Non-JSON output is passed through
If stdout of an `output: json` command is not valid JSON, the system SHALL write it unchanged to stdout and log an info-level message (shown with `--verbose`), and SHALL NOT show a warning. stderr SHALL always be passed through unchanged and uncaptured. When the arguments contain `-h` or `--help` (before `--`), the system SHALL NOT capture or process the output.

#### Scenario: Invalid JSON
- **WHEN** the binary prints `hello` to stdout
- **THEN** `hello` is printed without a warning
- **AND** running with `--verbose` logs that the output was not JSON

#### Scenario: Help of a JSON command
- **WHEN** the user runs `c8y foo services list --help` for a command with `output: json`
- **THEN** the binary's help is printed unchanged and no message about JSON is shown

### Requirement: Raw output commands are not processed
Commands without `output: json` (the default `raw`), passthrough mounts, and script commands SHALL have stdout connected directly, with no buffering or processing, and SHALL receive all arguments unchanged.

#### Scenario: Interactive binary
- **WHEN** a raw binary command prompts the user on stdout and reads stdin
- **THEN** the prompt appears immediately and the input is read

