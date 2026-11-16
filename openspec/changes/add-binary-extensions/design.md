## Context

go-c8y-cli's extension system was modeled on `gh extension`. Today an extension is a git repository (or a local folder) installed under `settings.extensions.datadir`, containing any of:

- `commands/`: executable files, one per command (scripts)
- `api/`: declarative command specs (YAML, parsed by `pkg/cmdparser`)
- `views/`, `templates/`: output views and jsonnet templates
- `extension.yaml`: aliases

The binary path from gh (`installBin`, `isBinExtension`, `manifest.yml`, Rosetta fallback, arm64 codesign) is present but gated behind `Manager.allowBinary`, which is never set. Problems in that path:

1. Only a bare binary is downloaded to `<datadir>/<name>/<name>`; nothing else from the repository is installed.
2. `parseBinaryExtensionDir` sets `path` to the binary file, so `Commands()`, `ViewPath()`, `TemplatePath()`, and `Aliases()` join paths onto a file.
3. `ConvertToCobraCommands` and `Manager.Dispatch` only find commands under `commands/`, so a binary extension registers nothing.
4. `root.go` always calls `Execute(..., isBinary=false, ...)`, which wraps the binary in `sh` on Windows.
5. Extension processes inherit the environment only. `--session`, `--dry`, and other global options are not passed on.

This change depends on `unify-extension-command-tree`, which introduces `pkg/extensions/cmdtree`, a source-neutral tree with conflict precedence.

## Goals / Non-Goals

**Goals:**
- A compiled binary can provide commands, and the same extension can also provide views, templates, API specs, scripts, and aliases.
- Install, upgrade, pin, and remove work the same for binary extensions as for git extensions.
- Binary commands feel native: help, tab completion, output formatting, and session handling.
- A small, stable Go SDK and a project template so authors don't start from scratch.
- Ship as a single PR (see Migration Plan for split points if review requires it).

**Non-Goals:**
- Release hosts other than GitHub and GitHub Enterprise for binary installs. Other hosts keep using `git clone`.
- Signature verification (Sigstore, GPG). Checksums only for now.
- Sandboxing extension processes.
- SDKs for languages other than Go. The runtime context and JSON output contract are language neutral, so other SDKs can follow.
- Running extensions in-process (Go plugins, WASM).

## Decisions

### 1. Release asset format: archive first, bare binary as fallback

Asset selection on a GitHub release, in order:

1. Archive matching the platform: `<anything>[_-]<os>[_-]<arch>.tar.gz` (or `.zip`, typically for Windows), case-insensitive, where `<arch>` uses Go names (`amd64`, `arm64`, ...). This matches GoReleaser's default `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}`.
2. Bare binary: `<anything>-<os>-<arch>[.exe]` (current gh convention, current `installBin` matching).
3. On darwin-arm64 with Rosetta, the same two rules for darwin-amd64 (existing behavior).

Archive contents:

```
[c8y-foo/]                  # optional single top-level folder, stripped
  extension.yaml            # required in an archive
  bin/c8y-foo[.exe]         # or the path set in extension.yaml binary.path
  views/ templates/ api/ commands/   # optional
```

*Why archives:* a bare binary can't carry views, templates, or API specs, which is the main gap. *Alternative considered:* install the git repository and download the binary into it. That needs git on the machine, mixes two version sources (commit and release tag), and breaks when the binary and the repository files get out of sync. *Alternative considered:* embed assets in the binary and extract them on first run. That hides files from c8y's view and template lookup and requires every author to implement it.

### 2. Install layout and model

```
<datadir>/c8y-foo/
  manifest.yml              # owner, name, host, tag, isPinned, asset name, sha256
  extension.yaml
  bin/c8y-foo
  views/ templates/ api/
```

- `Extension.Path()` returns the directory for every kind.
- New `Extension.Executable() string` returns the absolute binary path (empty for extensions without a binary). It is resolved from `extension.yaml` `binary.path`, defaulting to `bin/<dir-name>[.exe]`. For a bare-binary install, the installer writes a minimal `extension.yaml` with `binary.path: c8y-foo`.
- New `Extension.Kind()`: `git`, `local`, or `binary` (install source), independent of whether the extension has a binary. A local or git extension can also contain a binary (for example a developer running `go build -o bin/c8y-foo` and `c8y extensions install .`).
- Installs extract into a temporary folder next to the target and rename it into place, so a failed upgrade leaves the previous version intact.

### 3. `extension.yaml` additions

```yaml
version: v1
binary:
  path: bin/c8y-foo        # relative to the extension dir, ".exe" added on Windows if missing
  mount: tools             # command path under `c8y foo`; "" (default) means the extension root
  completion: cobra        # cobra | none (default none)
commands:
  - name: services list    # path under `c8y foo`
    description: List services
    exec: [services, list] # args passed to the binary; default is the name split on spaces
    output: json           # json | raw (default raw)
aliases: [...]             # unchanged
```

`commands` entries without a `binary` section are invalid (a warning is logged and the entry is ignored).

### 4. Command tree integration

Binary entries become two new `cmdtree` sources:

- **Declared commands** (`commands` list): leaf nodes, `SourceBinary`.
- **Mount** (`binary.mount`): a passthrough node. Arguments after the mount path are forwarded to the binary unchanged.

Updated leaf precedence: **API spec, then declared binary command, then script**. The mount is a fallback, not a leaf: if the mount path is also a group (because other sources define children there), known children run normally and any other arguments go to the binary. This uses cobra's behavior for non-root commands that have both `RunE` and subcommands (unknown positional args reach `RunE`), and needs a test to lock it in.

*Why not pass the whole extension namespace to the binary by default when other sources exist:* that would make API and script commands in the same extension unreachable, the same class of bug as #585. With the fallback approach, a binary mounted at the extension root coexists with API and script commands.

### 5. Execution

- Binaries run directly on all platforms (`Execute(..., isBinary=true)`); `sh` wrapping on Windows stays for scripts only.
- Flag parsing stays disabled for extension commands (except during `__complete`), so `--help`, `-h`, and `--examples` reach the binary.
- The exit code of the binary becomes the exit code of `c8y` (as `Dispatch` already does through `cmderrors.NewUserErrorWithExitCode`). The cobra `RunE` path in `root.go` is aligned with that.
- stdin, stdout, and stderr are connected directly, except for `output: json` commands (decision 7).

### 6. Runtime context (all extension commands)

c8y builds the child environment as: inherited environment, then the session variables from `c8ysession.GetVariablesFromSession(session, cfg, client, setPassword=false)` (the same set that `c8y sessions set` and `c8y sessions login` export to the shell), then:

| Variable | Value |
|---|---|
| `C8Y_EXTENSION_NAME` | name without `c8y-` |
| `C8Y_EXTENSION_DIR` | `Extension.Path()` |
| `C8Y_BIN` | `os.Executable()` of the running c8y |
| `C8Y_SETTINGS_DEFAULTS_DRY` | `true` when `--dry` is active |
| `C8Y_SETTINGS_MODE`, `C8Y_MODE` | effective session mode (respects `--sessionMode`) |
| `C8Y_SETTINGS_DEFAULTS_INSECURE` | `true` when TLS verification is turned off (`--insecure`) |
| `C8Y_EXTENSION_CONTEXT_VERSION` | `1`, so SDKs can detect what is available |

c8y does not force a login before running an extension. It passes the credentials it already has (stored token, or the authorization header for basic auth). If the token has expired, the SDK returns an error telling the user to run `set-session` or `c8y sessions login`.

*Alternative considered:* the binary calls `c8y` to get credentials (`$C8Y_BIN sessions get`). This works from any language, but it adds a process spawn per command and duplicates c8y's decryption prompt handling. It's still available to authors as a fallback through `C8Y_BIN`.

### 7. JSON output processing

For commands with `output: json`:

1. Before forwarding, c8y removes its own output flags from the args: `--output/-o`, `--select`, `--filter`, `--view`, `--compact/-c`, `--raw/-r`, `--noColor/-M`. Binaries cannot define flags with these names on JSON commands (documented).
2. c8y captures stdout, reads it as a JSON array, a JSON object, or newline-delimited JSON objects, and passes each value through `Factory.WriteOutput`, which applies views, selection, filtering, and formatting.
3. stderr is passed through unchanged. If stdout is not valid JSON, c8y writes it unchanged and logs an info message (text output such as messages is valid, so no warning).
4. With `-h` or `--help`, the output is not captured, so the binary's help is shown directly.

*Why host-side:* formatting lives inside the CLI (`cmdutil`, `dataview`, `console`). Re-implementing it in an SDK would drift, and a host-side pipeline works for extensions written in any language.

### 8. Go SDK as a nested module

`pkg/extsdk` with its own `go.mod` (module `github.com/reubenmiller/go-c8y-cli/pkg/extsdk`, tags `pkg/extsdk/vX.Y.Z`), depending only on `go-c8y` and `cobra`. Importing the main CLI module would pull in its full dependency graph and tie extension builds to CLI internals.

The module path has no `/v2`: the root module is `github.com/reubenmiller/go-c8y-cli/v2` (major version suffix, no `v2/` folder), so the Go tooling resolves `.../v2/pkg/extsdk` as a package of the root module, which the nested `go.mod` excludes. A nested module uses its folder path within the repository instead. The SDK does not import the CLI module, so no `replace` directive or `go.work` file is needed for local development.

API surface (v0):

```go
ctx, err := extsdk.FromEnv()          // runtime context (decision 6)
client, err := ctx.Client()           // *c8y.Client configured from the context
err = extsdk.CheckError(err)          // adds a login hint to 401 errors
root := extsdk.NewRootCommand("foo")  // cobra root, hidden __complete support
extsdk.Execute(root)                  // runs the root command, sets the exit code
extsdk.Print(v)                       // writes NDJSON to stdout for output: json
extsdk.Exec(ctx, "inventory", "get", "--id", id) // runs $C8Y_BIN with the same env
ctx.ExtensionFile("templates/x.jsonnet")
ctx.DryRun()
```

*Alternative considered:* a separate repository (`go-c8y-cli-extsdk`). It's easier to version independently but harder to keep in sync with the runtime context contract. Nested module keeps both in one PR. It can move out later without breaking importers if a vanity import path is used from the start. (Open question.)

### 9. Scaffolding

`c8y extensions create <name> --precompiled go`:
- `main.go` and `cmd/` using `extsdk`, with one example `output: json` command
- `extension.yaml` with `binary` and `commands`, `views/` with an example view
- `.goreleaser.yaml`: builds `bin/<name>`, archives with `extension.yaml`, `views/`, `templates/`, `api/` included, produces `checksums.txt`
- `.github/workflows/release.yml`: runs GoReleaser on tag push

`--precompiled other`: the same layout without Go files, with a `script/build.sh` placeholder and a release workflow that expects the author to produce the archives.

### 10. System-wide extension folders

Package managers (Homebrew, apt/rpm, Scoop) install files system-wide, but `settings.extensions.datadir` is per user, so a package can't run `c8y extensions install` for every user. c8y therefore also reads extensions from system-wide folders:

- `settings.extensions.systemDirs`: a list of folders separated by the OS path list separator. If not set, the defaults are used:
  - `<dir of c8y executable>/../share/c8y/extensions` (both the invoked path and the resolved symlink target), which covers Homebrew and most prefix-based installs
  - `/opt/homebrew/share/c8y/extensions`, `/home/linuxbrew/.linuxbrew/share/c8y/extensions`, `/usr/local/share/c8y/extensions`, `/usr/share/c8y/extensions`
  - `%ProgramData%\c8y\extensions` on Windows
- Folders that don't exist are ignored. Entries must start with `c8y-` and be folders (symlinks to folders are followed, as Homebrew links from the Cellar).
- Precedence on name conflicts: the user's extension folder, then system folders in order. Installing an extension into the user folder with the same name as a system extension is allowed, so users can override a system-wide version.
- System extensions have the kind `system`. c8y never writes to these folders: `update` and `delete` fail with a message to use the package manager. If the folder contains a `manifest.yml` (for example a packaged binary release), its tag is shown as the version.

*Alternative considered:* a package post-install hook that runs `c8y extensions install <dir>`. It only installs for the user running the package manager, and Homebrew discourages post-install steps that write to the home folder.

## Risks / Trade-offs

- [Running downloaded binaries] → Binary installs only come from an explicit `c8y extensions install`. Checksums from `checksums.txt` (GoReleaser format) are verified when present. Install output shows the repository, tag, and asset name. This is the same trust model as `gh`.
- [Archive extraction (zip-slip, symlinks)] → Reject entries with absolute paths, `..` segments, or symlinks that point outside the target folder. Limit the total extracted size.
- [Credentials in child process environment] → Same exposure as `set-session` today, which exports these variables to the user's shell. Passwords are not passed (`setPassword=false`). Document that extensions run with the user's session.
- [New env vars override values a script exported itself] → Only the documented variables are set. Note it in the changelog.
- [cobra fallback behavior for groups with `RunE`] → Covered by a unit test so a cobra upgrade that changes it fails the build.
- [Single PR is large] → Tasks are ordered so the PR can be split if review needs it: (a) model and install, (b) command tree and execution, (c) runtime context, (d) JSON output, (e) scaffolding and SDK.
- [System folders are writable by other users on misconfigured systems] → c8y only reads from them, and the default locations are the same ones package managers already use for executables. Users can set `settings.extensions.systemDirs` to a non-existent path to turn them off.
- [Startup cost of reading `extension.yaml` for every extension] → Already read for aliases. Parse once and cache on the `Extension` value.

## Migration Plan

- Existing git and local extensions need no changes.
- A git extension that later publishes binary release assets is migrated on `c8y extensions upgrade` (existing `upgradeExtension` logic, now enabled).
- Rollback: set `allowBinary` back to false, which keeps the model and runtime context changes but stops new binary installs.
- Release: tag `pkg/extsdk/v0.1.0` (`task release-extsdk -- v0.1.0`) alongside the CLI release that ships this change.

## Open Questions

- Should `c8y` force a login before running extensions that need the tenant (for example `binary.session: required`)? Proposed: not in v1; revisit after feedback.
- Should `output: json` commands also support c8y's `--includeAll` and paging flags, or leave paging to the binary?
- Allow non-GitHub release hosts (GitLab, generic URL) later through a `--from-url` install option?
