## ADDED Requirements

### Requirement: Name checks
The system SHALL report, under the check `name`: a declared `name` which doesn't match `^[a-z0-9][a-z0-9-]*$` after a leading `c8y-` is removed and the name is converted to lowercase (error), a declared name which is the name of a built-in command (error), and a declared name with the `c8y-` prefix or with uppercase letters (warning, with the name that is used). For an installed extension whose installed name differs from its declared name, the system SHALL report an info finding with both names.

#### Scenario: Invalid name
- **WHEN** `extension.yaml` has `name: My_Tool`
- **THEN** an error names the value and the allowed format

#### Scenario: Built-in command name
- **WHEN** `extension.yaml` has `name: devices`
- **THEN** an error says that `devices` is a built-in command

#### Scenario: Redundant prefix
- **WHEN** `extension.yaml` has `name: c8y-analytics`
- **THEN** a warning says to use `analytics`

#### Scenario: Uppercase letters
- **WHEN** `extension.yaml` has `name: Analytics`
- **THEN** a warning says that the extension is installed as `analytics`

#### Scenario: Installed under another name
- **WHEN** the extension was installed with `--name myanalytics` and its `extension.yaml` has `name: analytics`
- **THEN** an info finding says that the extension is installed as `myanalytics` and declares `analytics`
- **AND** `--strict` doesn't fail because of it
