## ADDED Requirements

### Requirement: Binary declaration in extension.yaml
An extension SHALL declare a binary with a `binary` section in `extension.yaml` containing `path` (relative to the extension folder), optional `mount` (command path under `c8y <ext>`, default the extension root), and optional `completion` (`cobra` or `none`, default `none`). On Windows, `.exe` SHALL be appended to `path` when it has no extension.

#### Scenario: Default path
- **WHEN** `extension.yaml` of `c8y-foo` has `binary: {}`
- **THEN** the binary is resolved to `bin/c8y-foo` (`bin/c8y-foo.exe` on Windows)

#### Scenario: Missing binary file
- **WHEN** `binary.path` points to a file that does not exist
- **THEN** a warning naming the extension and path is logged and no binary commands are registered for that extension
- **AND** the extension's other commands, views, and templates still work

### Requirement: Mount passthrough
The system SHALL register the binary at its mount path. All arguments after the mount path SHALL be passed to the binary unchanged, including `--help`, `-h`, and `--examples`.

#### Scenario: Root mount
- **WHEN** `c8y-foo` has `binary: {}` and no other command sources, and the user runs `c8y foo services list --pageSize 5`
- **THEN** the binary runs with args `services list --pageSize 5`

#### Scenario: Sub-path mount
- **WHEN** `binary.mount` is `tools` and the user runs `c8y foo tools check --verbose`
- **THEN** the binary runs with args `check --verbose`

#### Scenario: Help is forwarded
- **WHEN** the user runs `c8y foo --help` and the binary is mounted at the root with no other command sources
- **THEN** the binary's help output is shown

### Requirement: Mount coexists with other command sources
When the mount path is also a group that contains commands from other sources, the system SHALL run those commands for their paths and SHALL pass any other arguments under the mount path to the binary.

#### Scenario: Root mount with API commands
- **WHEN** `c8y-foo` has an API group `devices` with command `list`, and the binary is mounted at the root
- **THEN** `c8y foo devices list` runs the API command
- **AND** `c8y foo report --month 10` runs the binary with args `report --month 10`

### Requirement: Declared binary commands
Each entry in the top-level `commands` list of `extension.yaml` SHALL be registered at `c8y <ext> <name>` with its `description`, and SHALL run the binary with the entry's `exec` args (default: `name` split on spaces) followed by the user's arguments.

#### Scenario: Declared command
- **WHEN** `commands` has `{name: "services list", description: "List services"}`
- **THEN** `c8y foo services --help` lists `list` with description `List services`
- **AND** `c8y foo services list --name x` runs the binary with args `services list --name x`

#### Scenario: Custom exec args
- **WHEN** `commands` has `{name: "ls", exec: ["services", "list"]}`
- **THEN** `c8y foo ls` runs the binary with args `services list`

#### Scenario: Declared command without a binary
- **WHEN** `commands` entries exist but `extension.yaml` has no `binary` section
- **THEN** a warning is logged and the entries are ignored

### Requirement: Direct execution and exit codes
The system SHALL run extension binaries directly (not through `sh`) on every platform, connect stdin, stdout, and stderr (except as defined by `extension-json-output`), and exit with the binary's exit code.

#### Scenario: Non-zero exit
- **WHEN** the binary exits with code 3
- **THEN** `c8y` exits with code 3

#### Scenario: Windows
- **WHEN** a binary extension runs on Windows without Git for Windows installed
- **THEN** the command runs (no `sh.exe` is required)

### Requirement: Tab completion forwarding
When `binary.completion` is `cobra`, the system SHALL answer completion requests under the mount path or a declared command by running the binary with `__complete` followed by the arguments relative to the binary, and SHALL return its results. When `completion` is `none`, the binary SHALL NOT be run for completion.

#### Scenario: Completion forwarded
- **WHEN** `binary.completion: cobra` and the user presses tab after `c8y foo services list --na`
- **THEN** c8y runs the binary with `__complete services list --na` and shows its suggestions

#### Scenario: Completion disabled
- **WHEN** `binary.completion` is not set
- **THEN** completion under the binary's paths offers no binary-provided suggestions and the binary is not started
