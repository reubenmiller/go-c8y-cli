## Why

Extension commands come from two sources (API spec files under `api/` and executable files under `commands/`), and each source builds its own cobra command groups without checking the other. When both define the same group, cobra ends up with two sibling commands of the same name and resolves to the first one, so the other source's commands become unreachable ([#585](https://github.com/reubenmiller/go-c8y-cli/issues/585)). The same code also drops script commands nested more than two folders deep. Binary extensions (change `add-binary-extensions`) will add a third command source, so the merge logic needs to be correct and shared before that lands.

## What Changes

- Replace the per-source cobra wiring in `ConvertToCobraCommands` with a single extension command tree builder that merges all sources by command path.
- Command groups with the same path from different sources (API spec group, `commands/` folder, and later binaries) merge into one group.
- Script commands at any folder depth are registered under their full path.
- Two API spec files that declare the same `group.name` merge into one group instead of shadowing each other.
- When two sources define the same leaf command, a fixed precedence rule picks one and logs a message (shown with `--verbose`) that names both sources.
- Group descriptions from API specs take priority over generated placeholder descriptions (`"<name> command group"`).

## Capabilities

### New Capabilities
- `extension-command-tree`: How c8y builds the command tree for an installed extension from its command sources, including merging, nesting, conflict resolution, and help text.

### Modified Capabilities
<!-- None: no specs exist yet under openspec/specs/ -->

## Impact

- Code: `pkg/cmd/root/root.go` (`ConvertToCobraCommands`), new package for the tree builder (for example `pkg/extensions/cmdtree`), `pkg/cmdparser` (needs to expose the parsed group and commands separately, or return a group that can be merged).
- Behavior: commands that were unreachable become reachable. No commands that work today are removed. If an extension relied on an API command shadowing a script command of the same name, that still works, and a message is now logged with `--verbose`.
- Tests: new unit tests for the tree builder, and the `tests/testdata/extensions/c8y-kitchensink` fixture gains a script command inside an API-defined group.
- Fixes [#585](https://github.com/reubenmiller/go-c8y-cli/issues/585). Shipped as its own PR, ahead of binary extension support.
