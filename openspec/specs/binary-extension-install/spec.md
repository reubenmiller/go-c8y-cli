# binary-extension-install Specification

## Purpose
TBD - created by archiving change add-binary-extensions. Update Purpose after archive.
## Requirements
### Requirement: Binary install from GitHub releases
When installing an extension from a GitHub or GitHub Enterprise repository, the system SHALL install it as a binary extension if the selected release has an asset for the current platform, and SHALL otherwise install it with `git clone` as today.

#### Scenario: Release with platform archive
- **WHEN** the user runs `c8y extensions install owner/c8y-foo` and the latest release has `c8y-foo_1.2.0_linux_amd64.tar.gz` on a linux/amd64 machine
- **THEN** the archive is downloaded and extracted to `<datadir>/c8y-foo/`
- **AND** `c8y extensions list` shows `c8y-foo` with version `1.2.0` and kind `binary`

#### Scenario: Repository without releases
- **WHEN** the repository has no releases
- **THEN** the extension is installed with `git clone`, as before this change

#### Scenario: Release without an asset for the platform
- **WHEN** the latest release has binary assets but none for the current platform (and no Rosetta fallback applies)
- **THEN** the install fails with an error that names the repository and the missing platform

### Requirement: Platform asset selection
The system SHALL select a release asset in this order: an archive (`.tar.gz` or `.zip`) whose name ends with `<os>` and `<arch>` separated by `_` or `-` (case-insensitive, Go names); then a bare binary ending with `-<os>-<arch>` (plus `.exe` on Windows); then, on darwin-arm64 with Rosetta installed, the same rules for darwin-amd64.

#### Scenario: Archive preferred over bare binary
- **WHEN** a release has both `c8y-foo_darwin_arm64.tar.gz` and `c8y-foo-darwin-arm64`
- **THEN** the archive is installed

#### Scenario: Bare binary for gh compatibility
- **WHEN** a release has only `c8y-foo-windows-amd64.exe` and the machine is windows/amd64
- **THEN** the binary is installed to `<datadir>/c8y-foo/c8y-foo.exe`
- **AND** a generated `extension.yaml` sets `binary.path` to that file

#### Scenario: Rosetta fallback
- **WHEN** the machine is darwin/arm64 with Rosetta installed and the release only has darwin-amd64 assets
- **THEN** the darwin-amd64 asset is installed

### Requirement: Archive contents and layout
An archive SHALL contain `extension.yaml` at its root or inside a single top-level folder (which is stripped), plus the binary at `binary.path`. Any `views/`, `templates/`, `api/`, and `commands/` folders in the archive SHALL be installed and used the same way as for git extensions.

#### Scenario: Binary extension with views and templates
- **WHEN** the archive contains `extension.yaml`, `bin/c8y-foo`, `views/foo.json`, and `templates/foo.jsonnet`
- **THEN** the view is available to c8y output formatting
- **AND** the template can be referenced as `foo::foo.jsonnet` (the same syntax as for git extensions)

#### Scenario: Archive without extension.yaml
- **WHEN** the archive has no `extension.yaml`
- **THEN** the install fails and nothing is left in the data directory

### Requirement: Safe archive extraction
The system SHALL reject archive entries with absolute paths, `..` path segments, or symlinks that resolve outside the extension folder, and SHALL stop if the extracted size exceeds a fixed limit.

#### Scenario: Path traversal entry
- **WHEN** an archive contains an entry `../../.bashrc`
- **THEN** the install fails with an error naming the entry
- **AND** no file is written outside the temporary extraction folder

### Requirement: Checksum verification
When the release contains a `checksums.txt` asset (GoReleaser format), the system SHALL verify the SHA-256 of the downloaded asset against it and SHALL abort the install on a mismatch. The verified checksum SHALL be recorded in `manifest.yml`.

#### Scenario: Checksum mismatch
- **WHEN** the downloaded asset's SHA-256 does not match its entry in `checksums.txt`
- **THEN** the install fails and the previously installed version (if any) is unchanged

#### Scenario: No checksums published
- **WHEN** the release has no `checksums.txt`
- **THEN** the install continues and a notice says the asset was not verified

### Requirement: Atomic install and upgrade
The system SHALL extract into a temporary folder in the data directory and replace the extension folder only after download, verification, and extraction succeed.

#### Scenario: Failed upgrade
- **WHEN** an upgrade of `c8y-foo` fails during download
- **THEN** the previously installed version of `c8y-foo` still works

### Requirement: Pinning, upgrading, and removing binary extensions
The system SHALL support `--pin <tag>` to install a specific release, SHALL skip pinned extensions on upgrade, SHALL upgrade unpinned binary extensions to the latest release, and SHALL remove the whole extension folder on delete.

#### Scenario: Pinned install
- **WHEN** the user runs `c8y extensions install owner/c8y-foo --pin v1.0.0`
- **THEN** release `v1.0.0` is installed and `c8y extensions update c8y-foo` reports that pinned extensions cannot be updated

#### Scenario: Git extension migrates to binary
- **WHEN** a git-installed extension's repository starts publishing platform assets and the user runs `c8y extensions update`
- **THEN** the git checkout is replaced by the binary install of the latest release

### Requirement: Dry run
With `--dry`, install and upgrade SHALL report the release, asset name, and target folder without downloading or writing files.

#### Scenario: Dry run install
- **WHEN** the user runs `c8y extensions install owner/c8y-foo --dry`
- **THEN** the output names the asset that would be installed and the data directory is unchanged

### Requirement: Local extensions may contain a binary
A local extension (`c8y extensions install <dir>`) or git extension SHALL be able to declare a binary in `extension.yaml`, which is used in place without any download.

#### Scenario: Developer workflow
- **WHEN** a developer runs `go build -o bin/c8y-foo` in a folder whose `extension.yaml` sets `binary.path: bin/c8y-foo`, then `c8y extensions install .`
- **THEN** `c8y foo ...` runs the freshly built binary

### Requirement: System-wide extension folders
The system SHALL read extensions from the system-wide folders in `settings.extensions.systemDirs` (a list separated by the OS path list separator), or from the default folders when the setting is empty: `share/c8y/extensions` relative to the c8y executable, `/opt/homebrew/share/c8y/extensions`, `/home/linuxbrew/.linuxbrew/share/c8y/extensions`, `/usr/local/share/c8y/extensions`, and `/usr/share/c8y/extensions`, or `%ProgramData%\c8y\extensions` on Windows. Folders that do not exist SHALL be ignored. An extension in the user's extension folder SHALL take precedence over a system extension with the same name, and earlier system folders SHALL take precedence over later ones.

#### Scenario: Extension installed by a package manager
- **WHEN** a Homebrew formula installs `c8y-foo` (with `extension.yaml`, `bin/c8y-foo`, and `views/`) to `$(brew --prefix)/share/c8y/extensions/c8y-foo`
- **THEN** `c8y foo ...` runs the extension without running `c8y extensions install`
- **AND** `c8y extensions list` shows `c8y-foo` with kind `system`

#### Scenario: User extension overrides a system extension
- **WHEN** `c8y-foo` exists in a system folder and the user installs `c8y-foo` into their extension folder
- **THEN** the user's copy is used

### Requirement: System extensions are read-only
The system SHALL NOT update or remove extensions in system-wide folders, and SHALL tell the user to use the package manager instead.

#### Scenario: Update a system extension
- **WHEN** the user runs `c8y extensions update foo` and `c8y-foo` is a system extension
- **THEN** the command fails with a message that the package manager which installed it must update it
- **AND** the folder is unchanged

#### Scenario: Remove a system extension
- **WHEN** the user runs `c8y extensions delete foo` and `c8y-foo` is a system extension
- **THEN** the command fails with a message naming the folder and the package manager

