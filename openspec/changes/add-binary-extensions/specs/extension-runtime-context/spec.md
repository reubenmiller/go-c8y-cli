## ADDED Requirements

### Requirement: Session variables are passed to extension commands
For every extension command (script or binary, including passthrough mounts), the system SHALL add the session variables for the active session to the child process environment, using the same variable set that `c8y sessions set` exports (`C8Y_HOST`, `C8Y_TENANT`, `C8Y_USER`, `C8Y_TOKEN`, `C8Y_HEADER_AUTHORIZATION`, `C8Y_SESSION`, and the others), without the password.

#### Scenario: Session flag applies to script extensions
- **WHEN** the user runs `c8y --session prod.json foo list` and the script runs `c8y devices list`
- **THEN** the inner `c8y devices list` runs against the `prod.json` tenant

#### Scenario: Password not passed
- **WHEN** the active session uses basic authentication with a password stored in the session file
- **THEN** c8y does not add `C8Y_PASSWORD` to the child environment
- **AND** `C8Y_HEADER_AUTHORIZATION` is set
- **AND** a `C8Y_PASSWORD` value which the user already set in the environment is left unchanged

#### Scenario: No active session
- **WHEN** no session is active and the user runs a binary command
- **THEN** the binary still runs, and session variables that have no value are not set
- **AND** script commands keep their existing behavior of requiring an active session

### Requirement: Extension variables
The system SHALL set `C8Y_EXTENSION_NAME` (name without `c8y-`), `C8Y_EXTENSION_DIR` (extension folder), `C8Y_BIN` (absolute path of the running c8y executable), and `C8Y_EXTENSION_CONTEXT_VERSION` (`1`) for every extension command.

#### Scenario: Access to extension files
- **WHEN** an extension command reads `$C8Y_EXTENSION_DIR/templates/x.jsonnet`
- **THEN** it reads the file from the installed extension, for git, local, and binary installs

#### Scenario: Calling back into the same c8y
- **WHEN** an extension command runs `$C8Y_BIN devices list`
- **THEN** the same c8y executable that launched the extension is used, even if a different `c8y` is first in `PATH`

### Requirement: Global settings are passed
The system SHALL set `C8Y_SETTINGS_DEFAULTS_DRY=true` when dry run is active, `C8Y_SETTINGS_DEFAULTS_INSECURE=true` when TLS verification is turned off, and `C8Y_SETTINGS_MODE` and `C8Y_MODE` to the effective session mode for every extension command. c8y flags placed before the extension name (for example `c8y --session prod.json --dry foo list`) SHALL be applied by c8y and SHALL NOT be passed to the extension as arguments.

#### Scenario: Dry run propagates
- **WHEN** the user runs `c8y --dry foo cleanup` and the extension calls `c8y devices delete`
- **THEN** the inner command runs in dry run mode

#### Scenario: Session mode propagates
- **WHEN** the user runs `c8y --sessionMode prod foo cleanup`
- **THEN** `C8Y_SETTINGS_MODE` is `prod` in the child environment

### Requirement: No forced login
The system SHALL NOT perform a login before running an extension command. It SHALL pass the credentials that are already available.

#### Scenario: Offline extension command
- **WHEN** the user runs an extension command that does not contact Cumulocity and no tenant is reachable
- **THEN** the command runs without a login attempt
