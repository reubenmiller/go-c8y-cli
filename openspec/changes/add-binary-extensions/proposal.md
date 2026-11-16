## Why

Extensions can only provide commands as scripts (`commands/`) or declarative API specs (`api/`). Script extensions need a shell (Git for Windows `sh.exe` on Windows), are slow for heavy logic, and can't reuse Go libraries such as `go-c8y`. Users want to ship compiled extensions, the way `gh` supports precompiled extensions. Most of gh's binary install code was ported into `pkg/cmd/extension/manager.go` but is disabled (`allowBinary` is never set) and does not work end to end: it downloads only a bare binary (no views, templates, API specs, or aliases), sets `Path()` to the binary file so every component lookup fails, and binary commands are never registered because command discovery only reads the `commands/` folder.

## What Changes

- **Install from release archives.** `c8y extensions install <owner/repo>` detects a GitHub release asset for the current platform. It installs an archive (`.tar.gz` or `.zip`) containing the binary plus `extension.yaml`, `views/`, `templates/`, and `api/`, or a bare binary for gh compatibility. Checksums are verified when the release publishes them. Repositories without a matching asset still install through `git clone`.
- **System-wide extension folders for package managers.** c8y also reads extensions from system-wide folders (by default `share/c8y/extensions` next to the c8y executable, the Homebrew and Linuxbrew prefixes, `/usr/local/share` and `/usr/share`, and `%ProgramData%\c8y\extensions` on Windows), configurable with `settings.extensions.systemDirs`. Homebrew formulas, deb/rpm packages, or Scoop manifests can install an extension by placing its folder there. These extensions are read-only for c8y: the package manager updates and removes them.
- **One extension model for all kinds.** `Extension.Path()` is always the extension directory. A new `Executable()` returns the binary path. Views, templates, API specs, and aliases work the same way for git, local, and binary extensions.
- **Binary commands in the extension command tree.** Builds on `unify-extension-command-tree`:
  - A `binary` section in `extension.yaml` mounts the binary at a command path (the extension root by default). Arguments under the mount pass through to the binary.
  - An optional top-level `commands` list in `extension.yaml` declares individual binary commands with descriptions. These show in help and can mix with API and script commands in the same group.
  - Optional tab completion is forwarded to binaries built with cobra (`__complete`).
- **Runtime context for all extension commands.** c8y passes the active session (host, tenant, credentials as already produced by `c8ysession.GetVariablesFromSession`), the extension directory, the path to the running `c8y`, and global settings such as dry run, as environment variables. This also fixes `c8y --session X <ext> ...` for script extensions, which today ignore `--session`.
- **JSON output processing.** Commands declared with `output: json` write JSON to stdout. c8y applies its standard output pipeline (views, `--select`, `--filter`, `--output table|csv|json`), so binary commands format like built-in commands.
- **Scaffolding.** Re-enable `c8y extensions create --precompiled go|other`, which generates a Go project (or a language-neutral layout), a GoReleaser config that bundles the extension assets into the release archives, and a GitHub Actions release workflow.
- **Go extension SDK.** A small Go module (`pkg/extsdk`, its own `go.mod`) that reads the runtime context, builds a `go-c8y` client, provides a cobra root with the common flags, and writes output in the format the JSON pipeline expects.
- Turn on binary installs (`allowBinary`) and run binaries directly on Windows instead of through `sh`.

## Capabilities

### New Capabilities
- `binary-extension-install`: Installing, upgrading, pinning, listing, and removing binary extensions from GitHub releases, including archive layout, platform matching, and checksum verification.
- `binary-extension-commands`: How binary commands are declared in `extension.yaml`, mounted in the command tree, executed, and completed.
- `extension-runtime-context`: Environment variables c8y passes to every extension command (script and binary).
- `extension-json-output`: Host-side output processing for extension commands that produce JSON.
- `extension-scaffolding`: `c8y extensions create --precompiled` project templates and release tooling.
- `extension-go-sdk`: The Go SDK for writing binary extensions.

### Modified Capabilities
- `extension-command-tree`: Adds binary-declared commands and binary passthrough mounts as command sources, with their place in the conflict precedence. Depends on `unify-extension-command-tree` being merged and archived first.

## Impact

- Code: `pkg/cmd/extension` (manager, extension model, install and create commands, templates), `pkg/extensions` (interface, adds `Executable()` and `Kind()`), `pkg/extensions/cmdtree` (binary sources), `pkg/cmd/root/root.go`, `pkg/cmd/cmd.go` (dispatch), new `pkg/extsdk` module.
- Schemas: `tools/schema/extension.json` gains `binary` and `commands`.
- Docs: `docs/go-c8y-cli/docs/concepts/extensions`, `docs/go-c8y-cli/docs/tutorials/extensions`, and the generated `c8y extensions create` and `install` command docs.
- Security: c8y now downloads and runs binaries from GitHub releases, and passes session credentials to extension processes through the environment. See design.md.
- Compatibility: existing git and local script extensions keep working. The new environment variables are added to the inherited environment, so a script that already exports its own `C8Y_*` values sees the values for the active c8y session instead.
- Settings: new `settings.extensions.systemDirs` setting (`C8Y_SETTINGS_EXTENSIONS_SYSTEMDIRS`).
- Release: a nested Go module needs its own tag prefix (`pkg/extsdk/vX.Y.Z`).
