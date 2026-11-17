## ADDED Requirements

### Requirement: Scaffolds declare the extension name
`c8y extensions create <name>` SHALL write `name` into the generated `extension.yaml`, with the value of `<name>` without the `c8y-` prefix, for the script template and for `--precompiled go` and `--precompiled other`.

#### Scenario: Generated name
- **WHEN** the user runs `c8y extensions create c8y-foo --precompiled go`
- **THEN** the generated `extension.yaml` contains `name: foo`
- **AND** `c8y extensions validate` reports no findings for the folder

#### Scenario: Name without the prefix
- **WHEN** the user runs `c8y extensions create foo`
- **THEN** the folder `c8y-foo` is created and its `extension.yaml` contains `name: foo`
