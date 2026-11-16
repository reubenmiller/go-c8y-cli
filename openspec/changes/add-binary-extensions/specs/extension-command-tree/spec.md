## MODIFIED Requirements

### Requirement: Leaf command conflicts are resolved by fixed precedence
When more than one source defines a runnable command at the same path within an extension, the system SHALL keep the command from the highest-precedence source, in the order API spec, then declared binary command, then script, and SHALL log an info-level message (shown with `--verbose`) that names the kept and dropped origins, relative to the extension folder. A binary mount is not a leaf and does not take part in this precedence.

#### Scenario: API command and script with the same path
- **WHEN** an API spec declares `management health` and a script exists at `commands/management/health`
- **THEN** `c8y <ext> management health` runs the API command
- **AND** running with `--verbose` logs a message that includes both the spec file path and the script path
- **AND** no message is shown without `--verbose`

#### Scenario: Declared binary command and script with the same path
- **WHEN** `extension.yaml` declares binary command `management logs` and a script exists at `commands/management/logs`
- **THEN** `c8y <ext> management logs` runs the binary
- **AND** running with `--verbose` logs a message naming `extension.yaml` and the script path

## ADDED Requirements

### Requirement: Binary commands are a command source
Declared binary commands and binary mounts SHALL be merged into the extension command tree with the same group merging rules as API spec and script commands.

#### Scenario: Binary, API, and script commands in one group
- **WHEN** `management health` comes from an API spec, `management logs` from a script, and `management export` is a declared binary command
- **THEN** all three run from `c8y <ext> management <cmd>` and `c8y <ext> management --help` lists one `management` group with all three
