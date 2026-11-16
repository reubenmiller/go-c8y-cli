# extension-go-sdk Specification

## Purpose
TBD - created by archiving change add-binary-extensions. Update Purpose after archive.
## Requirements
### Requirement: Separate lightweight module
The Go SDK SHALL be a separate Go module (`github.com/reubenmiller/go-c8y-cli/pkg/extsdk`, released with tags `pkg/extsdk/vX.Y.Z`) whose dependencies are limited to `go-c8y`, `cobra`, and the standard library, so extensions do not import the CLI module.

#### Scenario: Dependency footprint
- **WHEN** an extension imports only `extsdk`
- **THEN** its `go.mod` does not require `github.com/reubenmiller/go-c8y-cli/v2`

### Requirement: Read runtime context
`extsdk.FromEnv()` SHALL return the runtime context defined by `extension-runtime-context` (session values, extension name and folder, c8y path, dry run, session mode) and SHALL return an error when `C8Y_EXTENSION_CONTEXT_VERSION` is missing or newer than the SDK supports.

#### Scenario: Run outside c8y
- **WHEN** the binary is run directly, not through `c8y`
- **THEN** `FromEnv()` returns an error explaining that the binary must be run as `c8y <ext> ...`

### Requirement: Cumulocity client
`ctx.Client()` SHALL return a `go-c8y` client configured from the context, using `C8Y_HEADER_AUTHORIZATION` if set, otherwise `C8Y_TOKEN`, otherwise username without password (which fails on first request with a clear error).

#### Scenario: Token session
- **WHEN** the context has `C8Y_HOST` and `C8Y_TOKEN`
- **THEN** requests from the client authenticate with that token against that host

#### Scenario: Expired credentials
- **WHEN** the first request returns 401
- **THEN** the returned error tells the user to run `c8y sessions login` or `set-session`

### Requirement: Output helper
`extsdk.Print(v)` SHALL write `v` as one JSON value per line to stdout, and SHALL write each element separately when `v` is a slice, so the output is valid input for `extension-json-output`.

#### Scenario: Slice output
- **WHEN** an extension calls `extsdk.Print([]Service{a, b})`
- **THEN** stdout contains two lines, each one JSON object

### Requirement: Calling c8y
`extsdk.Exec(ctx, args...)` SHALL run `C8Y_BIN` with the given args and the current environment, and return stdout, stderr, and the exit code.

#### Scenario: Reuse a built-in command
- **WHEN** an extension calls `extsdk.Exec(ctx, "devices", "get", "--id", "123")`
- **THEN** the same c8y executable and session are used

### Requirement: Root command helper
`extsdk.NewRootCommand(name)` SHALL return a cobra root command named after the extension, with cobra's completion command hidden and `__complete` enabled, so that `binary.completion: cobra` works without extra setup. `extsdk.Execute(root)` SHALL run the command, print an error to stderr, and exit with a non-zero exit code on failure.

#### Scenario: Completion works out of the box
- **WHEN** an extension built with `NewRootCommand` is installed with `binary.completion: cobra`
- **THEN** `c8y <ext> <tab>` lists the extension's subcommands

