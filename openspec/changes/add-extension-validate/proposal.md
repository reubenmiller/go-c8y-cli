## Why

Binary extensions (change `add-binary-extensions`) have rules that are easy to break without noticing: c8y handles its global flags after declared commands, so a binary flag such as `-v` or `--output` never reaches the binary; `commands` without a `binary` section are ignored; a missing binary only shows a warning; commands which conflict with another source are only reported with `--verbose`; and typos in `extension.yaml` keys are silently ignored. Extension authors need a way to find these problems before users hit them.

## What Changes

- New command `c8y extensions validate [<name> | <path>]` (default `.`) which checks an installed extension or an extension folder and lists the problems with a severity (`error`, `warning`, `info`), the check, the location, and a message.
- Checks:
  - manifest: invalid YAML, unknown keys (typos), invalid values (`binary.completion`, `commands[].output`), invalid or duplicate declared command names, `commands` without `binary`
  - binary: declared binary missing or not executable on this platform
  - flags: declared command flags which clash with c8y global flags, read from the binary with cobra's `__complete` protocol when `binary.completion` is `cobra` (otherwise an info message says that the flags were not checked)
  - command tree: commands dropped because another source (API spec, binary, script) defines the same path
  - API specs: files which can't be parsed
  - scripts: files under `commands/` which are not executable (not on Windows)
- Exit code: non-zero when there are errors, or warnings with `--strict`, so it can run in CI.
- `c8y extensions install` runs the validation after installing and prints a short summary when there are problems. The install itself is not changed.
- The `--precompiled go` template validates the extension in its release workflow.

## Capabilities

### New Capabilities
- `extension-validate`: Validating extensions, the checks, their severities, the output, and the exit code.

### Modified Capabilities
<!-- None: add-binary-extensions is not archived yet, and the install summary is additive. -->

## Impact

- Code: `pkg/cmd/extension` (new `validate` command and validator, install summary, template workflow).
- Docs: binary commands page (validation section), generated `c8y extensions validate` docs.
- Depends on `add-binary-extensions` (PR #743). Session provider checks for `add-extension-session-providers` (PR #744) are added in whichever of the two PRs merges second.
