## ADDED Requirements

### Requirement: Extensions declare session providers
An extension SHALL be able to declare session providers in the `sessionProviders` list of `extension.yaml`. Each provider SHALL have a `name` (matching `[a-z][a-z0-9-]*`), and MAY have a `description`, `exec` (arguments passed to the binary before the provider flags, default none), `secrets` (names of environment variables the provider needs), and `format` (`json`, `dotenv`, or `auto`, default `auto`). Providers SHALL only be available when the manifest also declares a `binary`.

#### Scenario: Provider declared with a binary
- **WHEN** extension `c8y-session-bitwarden` declares `binary: {path: bin/c8y-session-bitwarden}` and a provider `{name: bitwarden, exec: [list], secrets: [BW_SESSION]}`
- **THEN** the provider `bitwarden` is available to `c8y sessions login`

#### Scenario: Provider declared without a binary
- **WHEN** an extension declares `sessionProviders` but no `binary` section
- **THEN** its providers are not available
- **AND** a warning names the extension

#### Scenario: Duplicate provider name in one extension
- **WHEN** an extension declares two providers named `bitwarden`
- **THEN** the first one is used and a warning is logged

### Requirement: Providers are referenced by name
The system SHALL accept a provider reference as `<provider>` or `<extension>/<provider>` (extension name without `c8y-`). A short reference SHALL only resolve when exactly one installed extension declares a provider with that name.

#### Scenario: Unique short name
- **WHEN** only `c8y-session-bitwarden` declares a provider `bitwarden`
- **THEN** `bitwarden` and `session-bitwarden/bitwarden` both refer to it

#### Scenario: Ambiguous short name
- **WHEN** two installed extensions both declare a provider named `vault`
- **THEN** using `vault` fails with an error that lists `<extension-a>/vault` and `<extension-b>/vault`

#### Scenario: Unknown provider
- **WHEN** the user selects a provider which no installed extension declares
- **THEN** the login fails with an error that lists the available provider references

### Requirement: Select an extension provider
The system SHALL provide the session provider type `extension`, the `c8y sessions login --from-extension <ref>` flag, and the `session.provider.extension` setting. `--from-extension` SHALL select the `extension` type and override the configured provider. Without a `--from-*` flag, the type `extension` SHALL use the provider from `session.provider.extension`. Setting `session.provider.extension` SHALL NOT change the provider type by itself.

#### Scenario: Provider from the command line
- **WHEN** the user runs `c8y sessions login --from-extension bitwarden`
- **THEN** the `bitwarden` provider selects the session, regardless of `session.provider.type`

#### Scenario: Provider from settings for set-session
- **WHEN** `session.provider.type` is `extension` and `session.provider.extension` is `bitwarden`, and the user runs `set-session`
- **THEN** the `bitwarden` provider selects the session

#### Scenario: Setting alone does not change the type
- **WHEN** `session.provider.type` is `external` and `session.provider.extension` is `bitwarden`
- **THEN** `set-session` runs `session.provider.command` as before

#### Scenario: Extension type without a provider
- **WHEN** `session.provider.type` is `extension` and `session.provider.extension` is empty
- **THEN** the login fails with a message to set `session.provider.extension` or use `--from-extension`

### Requirement: Run the provider binary directly
The system SHALL run the extension's binary directly (not through `c8y` or `sh`) with the arguments: the provider's `exec`, then the same provider flags that `external` providers receive (`--clear`, `--loginType=<value>`, `--browserCallback=<value>`, `--sso-scopes=<value>`, `--sso-audience=<value>`, `-v`, `--debug`, when set), then the positional arguments of `c8y sessions login`. stdin and stderr SHALL be connected to the terminal, and stdout SHALL be parsed as the selected session using `--format`, otherwise the provider's `format`, otherwise automatic detection.

#### Scenario: Provider arguments
- **WHEN** the user runs `c8y sessions login --from-extension bitwarden --clear -v --folder c8y` and the provider declares `exec: [list]`
- **THEN** the binary runs with the args `list --clear -v --folder c8y`

#### Scenario: Interactive provider
- **WHEN** the provider shows an interactive picker on stderr and reads the choice from stdin
- **THEN** the picker is shown and the selected session from stdout is used

#### Scenario: Provider fails
- **WHEN** the provider exits with a non-zero exit code
- **THEN** the login fails, and no session is changed

#### Scenario: Binary missing
- **WHEN** the extension's binary does not exist
- **THEN** the login fails with an error naming the extension and the expected binary path

### Requirement: Provider environment
The system SHALL run the provider with the current environment without the session variables (`C8Y_HOST`, `C8Y_TENANT`, `C8Y_USER`, `C8Y_PASSWORD`, `C8Y_TOKEN`, `C8Y_HEADER_AUTHORIZATION`, `C8Y_SESSION`, and the other variables that `c8y sessions set` exports), plus `C8Y_EXTENSION_NAME`, `C8Y_EXTENSION_DIR`, `C8Y_BIN`, and `C8Y_EXTENSION_CONTEXT_VERSION`. The system SHALL add each secret from the provider's `secrets`, `session.provider.secrets`, and `--secrets`: from the environment if set, otherwise from the configured pin entry prompt.

#### Scenario: Declared secret from the environment
- **WHEN** the provider declares `secrets: [BW_SESSION]` and `BW_SESSION` is set
- **THEN** the provider receives `BW_SESSION` without any user configuration

#### Scenario: Declared secret from the pin entry
- **WHEN** the provider declares `secrets: [BW_SESSION]`, `BW_SESSION` is not set, and a pin entry command is configured
- **THEN** the value from the pin entry is passed to the provider as `BW_SESSION`

#### Scenario: Active session is not passed
- **WHEN** a session is active (for example `C8Y_HOST` and `C8Y_TOKEN` are set in the shell) and the user runs `set-session` with an extension provider
- **THEN** the provider's environment does not contain `C8Y_HOST` or `C8Y_TOKEN`

### Requirement: Provider discovery
The system SHALL complete provider references for `--from-extension` and for `c8y settings update session.provider.extension`, using the descriptions as completion help, and SHALL include the provider names of each extension in the output of `c8y extensions list`.

#### Scenario: Completion
- **WHEN** the user presses tab after `c8y sessions login --from-extension `
- **THEN** the installed providers are listed with their descriptions

#### Scenario: Extensions list
- **WHEN** the user runs `c8y extensions list --select name,sessionProviders`
- **THEN** `c8y-session-bitwarden` is shown with the provider `bitwarden`

### Requirement: Existing providers are unchanged
The provider types `auto`, `external`, `file`, `env`, `stdin`, and `interactive` SHALL keep their behavior, including the environment and arguments passed to `external` provider commands.

#### Scenario: External provider command
- **WHEN** `session.provider.type` is `external` and `session.provider.command` is `c8y session-bitwarden list --folder c8y`
- **THEN** the command runs as before, with the full current environment and the configured secrets
