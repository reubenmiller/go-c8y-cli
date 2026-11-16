## ADDED Requirements

### Requirement: Create a Go binary extension
`c8y extensions create <name> --precompiled go` SHALL create a git repository containing a Go module that uses the extension SDK, an `extension.yaml` with `binary` (with `completion: cobra`) and one declared `output: json` command, an example view, a `.goreleaser.yaml`, a GitHub Actions release workflow, and a README.

#### Scenario: Scaffold builds and runs locally
- **WHEN** the user runs `c8y extensions create c8y-foo --precompiled go`, then `go build -o bin/c8y-foo .` and `c8y extensions install .` in the new folder
- **THEN** the example command runs with `c8y foo <example-command>` and its output is formatted by c8y

#### Scenario: Release archives match the install format
- **WHEN** the generated GoReleaser config is run with `goreleaser release --snapshot`
- **THEN** it produces one archive per platform containing `extension.yaml`, `bin/c8y-foo`, and `views/`, plus `checksums.txt`, with names that `c8y extensions install` selects for each platform

### Requirement: Create a language-neutral binary extension
`c8y extensions create <name> --precompiled other` SHALL create the same layout without Go source files, with a build script placeholder and a release workflow that documents the expected archive names and contents.

#### Scenario: Other template
- **WHEN** the user runs `c8y extensions create c8y-foo --precompiled other`
- **THEN** the folder contains `extension.yaml` with a `binary` section, `script/build.sh`, a release workflow, and a README describing the archive format

### Requirement: Script extensions remain the default
Without `--precompiled`, `c8y extensions create` SHALL create the same script-based template as before this change.

#### Scenario: Default create
- **WHEN** the user runs `c8y extensions create c8y-foo`
- **THEN** the folder contains `commands/services/list` and no `binary` section

#### Scenario: Invalid precompiled value
- **WHEN** the user runs `c8y extensions create c8y-foo --precompiled rust`
- **THEN** the command fails with a message listing `go` and `other`
