## Context

Extension problems are found at different times today: unknown manifest keys never, conflicts only with `--verbose`, missing binaries as a warning on every command, and flag clashes only when a user's flag silently goes to c8y instead of the binary. The command tree (`pkg/extensions/cmdtree`) already reports conflicts through `OnConflict`, and cobra binaries list their flags with `<binary> __complete <args> -`.

## Goals / Non-Goals

**Goals:**
- One command which finds all known extension problems, for authors (local folder) and users (installed extension).
- Usable in CI (exit code, machine readable output).
- No network access and no Cumulocity session required.

**Non-Goals:**
- Running the extension's commands, or checking their output.
- Checking release assets on GitHub, or other platforms than the current one.
- Fixing problems automatically.

## Decisions

### 1. Command and target
`c8y extensions validate [<name> | <path>]`. If the argument is an existing folder, it is validated as a local extension (the folder name is the extension name); otherwise it is the name of an installed extension. Without an argument the current folder is used, which matches the development workflow (`c8y extensions install .`).

### 2. Findings and output
Each finding has `severity` (`error`, `warning`, `info`), `check` (`manifest`, `binary`, `flags`, `commands`, `api`, `scripts`), `location` (file or command path), and `message`. Findings are written with the c8y output handling (table on a terminal, json otherwise), so `--select` and `--output` work. Without findings, a success message is written to stderr.

Errors are problems which break the extension (the binary is missing, a file can't be parsed, commands are ignored). Warnings are problems which change behavior silently (flag clashes, conflicts, unknown keys). Info findings explain what was not checked.

Exit code: `101` (the c8y user error code; `1` is already used for HTTP 401) if there are errors, or warnings with `--strict`; `0` otherwise.

### 3. Flag clashes
For each declared command, when `binary.completion` is `cobra`, the binary runs with `__complete <exec args> -` (5 second timeout, extension variables in the environment, no session). Lines starting with `-` are flag names. A flag clashes if its long name or shorthand is a persistent flag of the c8y root command, except `--help`, `-h`, and `--examples`, which are still passed to the binary. The global flags come from the running c8y root command, so the check always matches the c8y version.

*Alternative considered:* a `flags` list in `extension.yaml`. It would work for non-cobra binaries, but it duplicates information which can be out of date, so it is not added for now.

### 4. Command tree conflicts
The validator builds a `cmdtree.Tree` with the same command paths as `ConvertToCobraCommands` (API spec groups and commands, declared binary commands and mount, scripts) and records the conflicts. Leaves are not built, so no cobra commands are created. A test runs the validator on the test extensions to keep the paths in sync with the real command registration.

*Alternative considered:* moving the command registration out of `pkg/cmd/root` so both use the same code. It is a larger refactoring of code under review in #743, and the path logic is small.

### 5. Manifest checks
`extension.yaml` is decoded a second time with `yaml.v3` `KnownFields(true)` to find unknown keys, which are reported as warnings with their line. Value checks mirror `tools/schema/extension.json`.

### 6. Install summary
After a successful install, `c8y extensions install` validates the extension without the flag check (which starts the binary) and, if there are errors or warnings, prints the number and the command to see them. The install result is not changed, so a problem in an extension never blocks installing it.

## Risks / Trade-offs

- [Path logic duplicated from command registration] → A test validates the kitchensink extension (API spec and script conflicts) and the SDK example, so differences show up as test failures.
- [Starting the binary during validation] → Only with `completion: cobra`, only the `__complete` command, with a timeout. The install summary does not start the binary.
- [Unknown key warnings for fields added in newer c8y versions] → Extensions written for a newer c8y get warnings on older versions, which is the useful signal.
