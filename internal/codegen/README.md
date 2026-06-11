# codegen — spec-driven CLI command generation

`internal/codegen` + `cmd/gen-cli` generate the cobra commands under
`pkg/cmd/**/*.auto.go` from the JSON specifications in `api/spec/json`.

This is a Go port of the PowerShell pipeline under `scripts/build-cli`
(`New-C8yApi.ps1`, `New-C8yApiGoRootCommand.ps1`, `New-C8yApiGoCommand.ps1`,
`New-C8yApiGoGetValueFromFlag.ps1`) and implements step 1 of
`go-c8y-v2/docs/proposals/CLI_CODE_GENERATION.md`.

`internal/codegen/powershell` + `cmd/gen-powershell` generate the PSc8y
module from the same specifications: the cmdlets under `tools/PSc8y/Public`,
the Pester tests under `tools/PSc8y/Tests` and the packaged module under
`tools/PSc8y/dist` (the port of `scripts/build-powershell` and
`tools/PSc8y/tools/build.psm1`).

## Usage

```sh
go run ./cmd/gen-cli            # regenerate pkg/cmd/**/*.auto.go (in place)
go run ./cmd/gen-cli -check     # verify outputs are current (CI-friendly)
task generate-go-code           # same as the first command, with task deps
task generate-go-code-legacy    # the original PowerShell pipeline
go test ./internal/codegen      # golden test: every spec vs committed output

go run ./cmd/gen-powershell           # regenerate + package the PSc8y module
go run ./cmd/gen-powershell -check    # verify cmdlets/tests are current
task build-powershell                 # same as the first command, with task deps
task build-powershell-legacy          # the original PowerShell pipeline
go test ./internal/codegen/powershell # golden test vs committed cmdlets/tests
```

## Byte-exact compatibility

The port reproduces the PowerShell output **byte-for-byte** for all 365
generated files. That includes deliberately ported quirks — do not "fix" these
without regenerating and reviewing the diff:

- PowerShell string semantics: `$true` interpolates as `True`, property access
  and comparisons are case-insensitive, single-element arrays unwrap, empty
  StringBuilders are truthy.
- `Get-C8yGoArgs` never receives `$UseOption`, so the shorthand-flag branches
  are dead code; unknown flag types yield a `nil` entry (no `cmd.Flags()`
  line) while their body setter is still emitted.
- The `||` fallback in the Cumulocity query builder
  (`$Properties.property || $Properties.name`) never falls back in PowerShell
  7 (an expression pipeline "succeeds" even when it yields `$null`), so only
  `property` is emitted.
- Formatting parity comes from running the same formatters as the build
  scripts: `go/format` (gofmt) for root commands and
  `golang.org/x/tools/imports` (goimports) for subcommands, which also
  resolves the conditionally-needed `pkg/c8ydata` import.

The PSc8y generator reproduces all committed cmdlets and tests byte-for-byte
as well, including these ported quirks:

- The `-replace` operator uses .NET substitution semantics on the
  *replacement* string: a test command containing `$_` expands to the entire
  template (see `Get-BulkOperationOperationCollection.auto.Tests.ps1`), which
  `dotnetReplace` reproduces.
- Commands sharing a `alias.powershell` name overwrite each other; the last
  command of a specification wins (e.g. `Remove-SoftwareVersion`).
- The `skipTest` flag of the *last* example applies to all test cases of a
  command, and stale files from removed commands are never deleted (only
  deprecated commands remove their files).
- Function lists and `.psm1` regions are ordered like `Get-ChildItem` sorts
  file names: case-insensitively.
- One intentional deviation: the dist `PSc8y.psd1` is version-stamped by
  substitution instead of `Update-ModuleManifest`, keeping the source
  formatting (the original rewrites the whole manifest). The result is
  semantically equivalent (verified with `Import-PowerShellDataFile` /
  `Test-ModuleManifest`), including the `CmdletsToExport = @()` and
  `DefaultCommandPrefix` normalizations.

## Roadmap (from CLI_CODE_GENERATION.md) — status

1. **Port the generator to Go** — ✅ done (this package).
2. **v2-service emitter** — ❌ superseded. `go-c8y-v2` has since adopted its
   own spec-driven generation (`tools/c8ygen`, see `go-c8y-v2/docs/API_GEN.md`):
   Layer 0 (`zz_generated_*.go` option structs, paths, enums, façade models)
   is generated from the *OpenAPI* spec + overlay, and the ergonomic service
   layer is deliberately hand-written, with a gating CI drift check. Emitting
   v2 services from the CLI spec would duplicate and conflict with that
   architecture. What remains useful from this step is vendoring/diffing: the
   CLI spec can be diffed against the OAS to find missing flags/endpoints.
3. **Thin-command emitter** (CLI commands calling v2 services + the
   `pkg/c8y/output` streaming pipeline, ~40 lines per command instead of
   ~200) — pending, with prerequisites:
   - A v2 `api.Client` in the CLI `cmdutil.Factory` (session config →
     `api.ClientOptions{BaseURL, Auth}`); today only login and the output
     packages use the v2 module. This belongs to the `feat/v2-output-streaming`
     integration track.
   - A **hand-written reference command** (e.g. `operations list`) proving the
     flag-parsing → `ListOptions` → `ListAll` iterator → `output.Render`
     shape end-to-end, including dry-run and the commander test suite. The
     emitter should only be written once that reference exists (the same
     prove-the-seam order `tools/c8ygen` used for its alarms pilot).
   - Migration is then group-by-group, with old and new command styles
     coexisting behind the same root command (the jsonfilter fallback
     pattern).
