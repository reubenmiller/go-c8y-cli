## 0. Prerequisite

- [x] 0.1 Confirm `unify-extension-command-tree` is merged and archived (provides `pkg/extensions/cmdtree` and the `extension-command-tree` spec)

## 1. Extension model

- [x] 1.1 Add `Executable() string` and `Kind()` to the `extensions.Extension` interface; regenerate mocks (`go generate ./pkg/extensions`)
- [x] 1.2 Make `Extension.Path()` return the extension folder for all kinds (fix `parseBinaryExtensionDir`)
- [x] 1.3 Parse `extension.yaml` once per extension and cache it (aliases, `binary`, `commands`); resolve `binary.path` with default `bin/<dir-name>` and `.exe` on Windows
- [x] 1.4 Add `binary` and `commands` to `tools/schema/extension.json`
- [x] 1.5 Unit tests for path resolution, missing binary warning, and `commands` without `binary`

## 2. Install, upgrade, remove

- [x] 2.1 Implement asset selection (archive by `<os>[_-]<arch>`, then bare binary, then Rosetta fallback) with table-driven tests over asset names
- [x] 2.2 Implement safe `.tar.gz` and `.zip` extraction (strip single top folder, reject absolute paths, `..`, escaping symlinks, size limit) with tests
- [x] 2.3 Download `checksums.txt` when present and verify SHA-256; store asset name and checksum in `manifest.yml`
- [x] 2.4 Extract to a temporary folder in the data dir and rename into place; keep the old version on failure
- [x] 2.5 Bare binary install: place binary in the extension folder and write a minimal `extension.yaml` with `binary.path`
- [x] 2.6 Require `extension.yaml` in archives; fail and clean up otherwise
- [x] 2.7 Keep arm64 codesign step for darwin binaries
- [x] 2.8 Respect `--dry` for install and upgrade (report release, asset, target)
- [x] 2.9 Enable `allowBinary` for GitHub and GitHub Enterprise hosts
- [x] 2.10 Show kind (`git`, `local`, `binary`) in `c8y extensions list`
- [x] 2.11 Tests with a fake GitHub release server (`httptest`) for install, pinned install, upgrade, git-to-binary migration, checksum mismatch, and missing platform
- [x] 2.12 Add `settings.extensions.systemDirs` with default package manager folders (`share/c8y/extensions` next to the executable, Homebrew, Linuxbrew, `/usr/local/share`, `/usr/share`, `%ProgramData%`)
- [x] 2.13 List system extensions (kind `system`, user folder wins on name conflicts, symlinks followed, `manifest.yml` version) and make them read-only for update and delete
- [x] 2.14 Tests for system folders: precedence, missing folders, symlinked folders, read-only update and delete

## 3. Command tree and execution

- [x] 3.1 Add `SourceBinary` declared leaves to `cmdtree` and update precedence to API, binary, script
- [x] 3.2 Add mount nodes: passthrough at the mount path, and fallback handling when the mount path is also a group
- [x] 3.3 Unit test that a non-root cobra command with subcommands and `RunE` receives unknown args (locks in cobra behavior used by the fallback)
- [x] 3.4 Run binaries directly on all platforms (`Execute(..., isBinary=true)`); keep `sh` wrapping for scripts on Windows
- [x] 3.5 Align exit code handling of the cobra `RunE` path with `Dispatch` (`cmderrors.NewUserErrorWithExitCode`)
- [x] 3.6 Update `Manager.Dispatch` to resolve binary paths through the same tree (or remove the duplicated lookup if cobra registration covers all cases)
- [x] 3.7 Forward completion to `<binary> __complete ...` when `binary.completion: cobra`; never start the binary otherwise
- [x] 3.8 Tests: root mount, sub-path mount, root mount with API commands, declared command with default and custom `exec`, help forwarding, exit code, completion forwarding

## 4. Runtime context

- [x] 4.1 Build the child environment in one helper used by every extension execution path: inherited env, `c8ysession.GetVariablesFromSession(..., setPassword=false)` (unset empty values), extension variables, dry run, session mode, `C8Y_EXTENSION_CONTEXT_VERSION=1`
- [x] 4.2 Do not trigger a login when building the context
- [x] 4.3 Tests: `--session` reaches a script extension, password not set, `--dry` and `--sessionMode` propagate, no session active

## 5. JSON output

- [x] 5.1 Strip c8y output flags (`-o/--output`, `--select`, `--filter`, `--view`, `-c/--compact`, `-r/--raw`, `-M/--noColor`) from args for `output: json` commands and apply them to the c8y output settings
- [x] 5.2 Capture stdout; parse array, object, or NDJSON; write each item through `Factory.WriteOutput`
- [x] 5.3 Pass invalid JSON through unchanged with a warning; never capture stderr
- [x] 5.4 Tests: table output with an extension view, `--select`, NDJSON, invalid JSON, raw command unaffected

## 6. Go SDK (`pkg/extsdk`, nested module)

- [x] 6.1 Create `pkg/extsdk/go.mod` (module `github.com/reubenmiller/go-c8y-cli/pkg/extsdk`) with only `go-c8y` and `cobra` (no `replace` or workspace needed, as the SDK does not import the CLI module)
- [x] 6.2 `FromEnv()` with context version check and a clear error when not run through c8y
- [x] 6.3 `ctx.Client()` (authorization header, then token, then username), 401 error message
- [x] 6.4 `Print(v)` NDJSON writer, `Exec(ctx, args...)`, `ctx.ExtensionFile(rel)`, `ctx.DryRun()`
- [x] 6.5 `NewRootCommand(name)` with hidden completion command
- [x] 6.6 Unit tests and a runnable example under `pkg/extsdk/examples`
- [x] 6.7 Add the module to CI (`go test` in `pkg/extsdk`) and document the `pkg/extsdk/vX.Y.Z` tag in the release process

## 7. Scaffolding

- [x] 7.1 Re-enable the `--precompiled go|other` flag on `c8y extensions create` (`pkg/cmd/extension/command.go`)
- [x] 7.2 Add Go templates under `pkg/cmd/extension/ext_tmpls/` (`main.go`, example command using `extsdk`, `extension.yaml` with `binary` and `commands`, example view, `.goreleaser.yaml`, release workflow, README)
- [x] 7.3 Add the `other` templates (`extension.yaml`, `script/build.sh`, release workflow, README describing archive names and contents)
- [x] 7.4 Test: scaffold, `go build -o bin/<name>`, `c8y extensions install .`, run the example command
- [x] 7.5 Test: `goreleaser release --snapshot --clean` on the scaffold produces archives that the asset selection picks for each platform (CI job, skipped if GoReleaser is not available)

## 8. Fixtures, integration tests, and docs

- [x] 8.1 Add a binary test extension fixture (`pkg/extsdk/examples/c8y-hello`, built by `task build-extension-fixtures` and loaded through `extensions.systemDirs`) that uses `extsdk`, a declared JSON command, a mount, a view, and an API spec in the same group
- [x] 8.2 Add integration tests in `tests/manual/extensions` for install from a local folder, command execution, output formatting, and session propagation
- [x] 8.3 Update `docs/go-c8y-cli/docs/concepts/extensions` (binary extensions, `extension.yaml` reference, runtime context variables, JSON output contract, system-wide folders and packaging for Homebrew/apt/Scoop, security notes)
- [x] 8.4 Add a tutorial `docs/go-c8y-cli/docs/tutorials/extensions/creating-a-binary-extension.md`
- [x] 8.5 Regenerate CLI docs for `c8y extensions create` and `install`
- [x] 8.6 Release notes in the PR description (CHANGELOG.md is not maintained), including the new environment variables passed to script extensions and the removed leading c8y flags
