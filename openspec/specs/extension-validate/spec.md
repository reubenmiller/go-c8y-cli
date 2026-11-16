# extension-validate Specification

## Purpose
TBD - created by archiving change add-extension-validate. Update Purpose after archive.
## Requirements
### Requirement: Validate command
The system SHALL provide `c8y extensions validate [<name> | <path>]`, which validates the extension in the given folder, or the installed extension with the given name, or the current folder when no argument is given. Each finding SHALL have a severity (`error`, `warning`, or `info`), a check name, a location, and a message, and SHALL be written using the c8y output handling.

#### Scenario: Validate a local folder
- **WHEN** the author runs `c8y extensions validate .` in an extension folder
- **THEN** the findings for that folder are listed

#### Scenario: Validate an installed extension
- **WHEN** the user runs `c8y extensions validate hello`
- **THEN** the findings for the installed extension `c8y-hello` are listed

#### Scenario: No problems
- **WHEN** the extension has no findings
- **THEN** nothing is written to stdout, a success message is written to stderr, and the exit code is 0

#### Scenario: Unknown extension
- **WHEN** the argument is neither a folder nor an installed extension
- **THEN** the command fails with an error naming the argument

### Requirement: Exit code
The command SHALL exit with code 101 (the c8y user error code) when there is at least one error, or at least one warning and `--strict` is set, and with code 0 otherwise.

#### Scenario: Warnings only
- **WHEN** the extension only has warnings
- **THEN** the exit code is 0, and 101 with `--strict`

### Requirement: Manifest checks
The system SHALL report: invalid YAML (error), unknown keys with their line (warning), invalid `binary.completion` or `commands[].output` values (error), declared commands with an empty or duplicate name (error), declared `commands` without a `binary` section (error), and session providers which are ignored because the manifest has no `binary` section, or because of an invalid or duplicate name or an invalid format (error).

#### Scenario: Typo in a key
- **WHEN** `extension.yaml` contains `comands:`
- **THEN** a warning names the key `comands` and its line

#### Scenario: Invalid session provider
- **WHEN** `extension.yaml` declares two session providers named `vault`
- **THEN** an error says that the duplicate session provider is ignored

#### Scenario: Commands without binary
- **WHEN** `extension.yaml` declares `commands` but no `binary`
- **THEN** an error says the commands are ignored without a binary

### Requirement: Binary checks
When a binary is declared, the system SHALL report an error if the file does not exist, is not a regular file, or (except on Windows) is not executable.

#### Scenario: Binary not built
- **WHEN** `binary.path` is `bin/c8y-hello` and the file does not exist
- **THEN** an error names the expected path

### Requirement: Flag clash checks
For each declared command, when `binary.completion` is `cobra`, the system SHALL read the command's flags from the binary (`__complete <exec args> -`) and report a warning for each flag whose long name or shorthand is a c8y global flag, except `--help`, `-h`, and `--examples`. The warning SHALL name the command, the flag, and the c8y flag, and suggest renaming the flag or using the mount. When `binary.completion` is not `cobra`, the system SHALL report an info finding that the flags were not checked.

#### Scenario: Shorthand clash
- **WHEN** the declared command `services list` has the flag `--verbose` with shorthand `-v`
- **THEN** a warning says that c8y handles `-v`/`--verbose`, so the binary never receives it

#### Scenario: No clash
- **WHEN** the declared command only has the flags `--status`, `--help`, and `-h`
- **THEN** no flag finding is reported for it

#### Scenario: Binary fails to complete
- **WHEN** running the binary with `__complete` fails or times out
- **THEN** a warning says that the flags of the command could not be read

### Requirement: Command tree checks
The system SHALL report a warning for each command which is dropped because another command source defines the same path, naming the kept and dropped origins.

#### Scenario: Script shadowed by an API command
- **WHEN** an API spec and a script both define `management health`
- **THEN** a warning says the script is ignored because the API command takes precedence

### Requirement: API spec and script checks
The system SHALL report an error for each API spec file which can't be parsed, and (except on Windows) a warning for each file under `commands/` which is not executable.

#### Scenario: Invalid API spec
- **WHEN** `api/devices.yaml` contains invalid YAML
- **THEN** an error names `api/devices.yaml`

### Requirement: Validation after install
After a successful `c8y extensions install`, the system SHALL validate the extension without starting its binary and, if there are errors or warnings, SHALL print their number and the command to show them on stderr. The install result SHALL NOT change.

#### Scenario: Install an extension with problems
- **WHEN** the user installs an extension with one error
- **THEN** the install succeeds, and stderr says that 1 problem was found and to run `c8y extensions validate <name>`

