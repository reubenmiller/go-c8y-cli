# extension-manifest-name Specification

## Purpose
How the name of an extension, and so its command `c8y <name>`, is chosen: the optional `name` in `extension.yaml`, its format, the order of precedence on install (`--name`, the declared name, the repository or folder name), the checks before an extension is moved into place, and the behavior on update.

## Requirements
### Requirement: Name in extension.yaml
`extension.yaml` SHALL support an optional top-level `name`: the extension name without the `c8y-` prefix, which is also the name of the extension's root command. The system SHALL remove a leading `c8y-` and convert the name to lowercase, and the result MUST match `^[a-z0-9][a-z0-9-]*$`. The JSON schema `tools/schema/extension.json` SHALL describe the property.

#### Scenario: Declared name
- **WHEN** the repository `owner/c8y-cli-extension-analytics` has `name: analytics` in `extension.yaml`, and the user runs `c8y extensions install owner/c8y-cli-extension-analytics`
- **THEN** the extension is installed to `<datadir>/c8y-analytics/`
- **AND** its commands are available as `c8y analytics ...`

#### Scenario: Name with the prefix
- **WHEN** `extension.yaml` has `name: c8y-analytics`
- **THEN** the extension is installed as `c8y-analytics`, the same as with `name: analytics`

#### Scenario: Name with uppercase letters
- **WHEN** `extension.yaml` has `name: Analytics`
- **THEN** the extension is installed as `c8y-analytics`, with the command `c8y analytics`

#### Scenario: Invalid name
- **WHEN** `extension.yaml` has `name: My Tool`
- **THEN** the install fails with an error that names the invalid value and the allowed format
- **AND** no extension folder is created

### Requirement: Name resolution on install
`c8y extensions install` SHALL use the first of: the `--name` flag, the `name` in the downloaded `extension.yaml`, and the repository name (or the folder name for local installs). This SHALL apply to git installs, binary installs from releases, and local installs.

#### Scenario: Flag overrides the declared name
- **WHEN** `extension.yaml` has `name: analytics` and the user runs `c8y extensions install owner/c8y-cli-extension-analytics --name myanalytics`
- **THEN** the extension is installed as `c8y-myanalytics`

#### Scenario: No declared name
- **WHEN** `extension.yaml` has no `name` and the user runs `c8y extensions install owner/c8y-foo`
- **THEN** the extension is installed as `c8y-foo`, as before this change

#### Scenario: Binary release with a declared name
- **WHEN** the release archive of `owner/c8y-cli-extension-analytics` contains `extension.yaml` with `name: analytics`
- **THEN** the binary extension is installed as `c8y-analytics`
- **AND** `c8y extensions update analytics` updates it from `owner/c8y-cli-extension-analytics`

#### Scenario: Local install with a declared name
- **WHEN** the folder `~/dev/my-analytics-tool` has `name: analytics` in `extension.yaml`, and the user runs `c8y extensions install .` in it
- **THEN** the extension is installed as `c8y-analytics`

#### Scenario: Reinstalling a local extension
- **WHEN** a local extension is installed as `c8y-analytics` and the user runs `c8y extensions install .` again in the same folder
- **THEN** the install succeeds and the extension still points to the folder

### Requirement: Name checks after download
The system SHALL download or clone remote extensions into a temporary folder in the extension directory, resolve the name, and check it before moving the folder into place. The install SHALL fail, and the temporary folder SHALL be removed, if the resolved name is the name of a built-in command or of another installed extension (other than an extension in a system folder). Local installs SHALL be checked in the same way.

#### Scenario: Declared name clashes with a built-in command
- **WHEN** `extension.yaml` has `name: devices`
- **THEN** the install fails with an error that `devices` matches the name of a built-in command
- **AND** no extension folder is left in the extension directory

#### Scenario: Declared name clashes with an installed extension
- **WHEN** `c8y-analytics` is installed and the user installs another repository whose `extension.yaml` has `name: analytics`
- **THEN** the install fails with an error that an installed extension already provides the `analytics` command
- **AND** the installed `c8y-analytics` is unchanged

#### Scenario: Git install of a pinned commit
- **WHEN** the user runs `c8y extensions install owner/c8y-cli-extension-analytics --pin abc1234` and the repository has no release assets for the platform
- **THEN** the repository is cloned, `abc1234` is checked out, and the folder is installed as `c8y-analytics` with the pin

### Requirement: Repository name prefix
The repository name SHALL only need to start with or contain `c8y-` when neither `--name` nor `name` in `extension.yaml` is set. The error SHALL say that the repository name must start with `c8y-` or that `extension.yaml` must set `name`.

#### Scenario: Repository without the prefix and with a declared name
- **WHEN** the repository `owner/analytics-tool` has `name: analytics` in `extension.yaml`
- **THEN** it is installed as `c8y-analytics`

#### Scenario: Repository without the prefix and without a name
- **WHEN** the repository `owner/analytics-tool` has no `name` in `extension.yaml`, and the user doesn't pass `--name`
- **THEN** the install fails with an error which says to start the repository name with `c8y-` or to set `name` in `extension.yaml`
- **AND** no extension folder is left in the extension directory

### Requirement: Updates keep the installed name
`c8y extensions update` SHALL NOT rename an installed extension. If the update changes the declared name (the `name` in `extension.yaml` before the update differs from the one after it) and the new name differs from the installed name, the system SHALL print a notice with the new name and the commands to reinstall the extension under it. Extensions installed with `--name` SHALL only get the notice when the declared name changes, not on every update.

#### Scenario: Declared name changes in a new version
- **WHEN** `c8y-analytics` is installed and the next release declares `name: dtm-analytics`
- **THEN** `c8y extensions update analytics` updates the extension, which is still available as `c8y analytics`
- **AND** a notice names `dtm-analytics` and shows the `c8y extensions delete` and `c8y extensions install` commands

#### Scenario: Extension installed with --name
- **WHEN** the extension was installed with `--name myanalytics`, declares `name: analytics`, and an update doesn't change the declared name
- **THEN** `c8y extensions update myanalytics` prints no name notice

### Requirement: Dry run
With `--dry`, the system SHALL NOT download the extension, and the message SHALL say that the name may be replaced by the name declared in `extension.yaml`.

#### Scenario: Dry run without --name
- **WHEN** the user runs `c8y extensions install owner/c8y-cli-extension-analytics --dry`
- **THEN** nothing is downloaded or installed
- **AND** the output says that the name is taken from `extension.yaml` if it declares one

