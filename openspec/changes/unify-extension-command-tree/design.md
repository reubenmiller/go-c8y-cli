## Context

`ConvertToCobraCommands` (`pkg/cmd/root/root.go`) registers each installed extension as `c8y <ext>`:

1. **API specs:** every `api/*.yaml` is passed to `cmdparser.ParseCommand`, which returns a fully built cobra group (`Use: spec.Group.Name`) that is added directly to the extension root. It is not recorded anywhere else.
2. **Scripts:** every file under `commands/` becomes a path such as `<ext> management logs`. Parent groups are looked up in a local `extCommandTree` map and created on a miss.

The two steps share no state:

- An API group `management` and a script folder `commands/management/` produce two sibling cobra commands named `management`. Cobra's `Find` returns the first match (the API group), so `c8y <ext> management logs` fails with "unknown command" (#585).
- For scripts nested deeper than `commands/<group>/<cmd>`, the intermediate group is only attached to the extension root when `len(parts) == 3`, so deeper groups are created but never attached.
- Two spec files with the same `group.name` produce two sibling groups.

There are no unit tests for this function today.

## Goals / Non-Goals

**Goals:**
- One place that turns the command sources of an extension into a cobra tree, with deterministic merging.
- Fix #585, the deep nesting bug, and duplicate spec groups.
- A source-neutral model, so binary extensions can add their commands as another source without changing the merge logic.
- Unit tests that do not need a Cumulocity tenant.

**Non-Goals:**
- Binary extension support (change `add-binary-extensions`).
- Changing the API spec format or the `commands/` folder layout.
- Conflicts between an extension and built-in c8y commands. Extensions stay namespaced under `c8y <ext>`.

## Decisions

### 1. Intermediate node model, built before any cobra command

```go
type Node struct {
    Name        string
    Short, Long string
    ShortIsDefault bool        // generated placeholder text, can be replaced
    Children    map[string]*Node
    Leaf        *Leaf          // nil for groups
}

type Leaf struct {
    Source  SourceKind         // SourceAPI, SourceScript (SourceBinary later)
    Origin  string             // path relative to the extension folder, for log messages
    Build   func() *cobra.Command
}
```

Each source adds entries with `tree.AddGroup(path, short, long)` and `tree.AddLeaf(path, leaf)`. Only after all sources are added does `tree.ToCobra()` create the cobra commands.

*Alternative considered:* merge directly on cobra commands by searching `cmd.Commands()` for an existing child before adding. This is a smaller diff, but it mixes merging with cobra construction, makes precedence harder to apply (the first registration already won), and is harder to test. The node model also gives the binary change a clean extension point.

### 2. API specs contribute a group plus leaves, not a finished cobra group

`cmdparser.ParseCommand` is split so the caller gets the group metadata and one builder per command (for example `cmdparser.ParseSpec(r) (*Spec, error)` and `cmdparser.NewCommand(spec, item, factory) *cobra.Command`). `ParseCommand` stays as a thin wrapper for any other callers.

### 3. Conflict precedence for leaves

When two sources define the same leaf path: **API spec, then script**. The binary change appends **binary** between them (API, binary, script). The losing leaf is dropped and an info-level message is logged with both origins, for example:

`Ignoring conflicting extension command. extension=analytics, path="analytics management health" kept=api/health.yaml dropped=commands/management/health`

API specs win because they are declarative and offer completion and validation. This also matches today's observable behavior, so no extension silently changes which command runs.

A path that is a leaf in one source and a group in another (for example a script file `commands/management` and an API group `management`) is a conflict where the group wins, and the leaf is dropped with an info-level message. A group cannot also be runnable in cobra without breaking subcommand lookup.

### 4. Group descriptions

A group created implicitly (from a script folder) gets the placeholder `"<name> command group"` and is marked `ShortIsDefault`. An explicit description from an API spec replaces a placeholder. If two specs give different explicit descriptions, the first one (in sorted file order) is kept.

### 5. Deterministic ordering

Spec files and script files are walked in lexical order (as `filepath.WalkDir` already does), and API specs are added before scripts. Results therefore do not depend on file system order.

### 6. Package location

Put the builder in `pkg/extensions/cmdtree` (no dependency on `pkg/cmd/root`), and have `ConvertToCobraCommands` call it. The executor callback for script leaves stays in `root.go`, passed in as a function, so the package does not need the full factory.

## Risks / Trade-offs

- [Spec parsing refactor touches `cmdparser`, which is shared with other code paths] → Keep `ParseCommand` behavior identical as a wrapper, and add a test that compares the old and new output for the kitchensink spec files.
- [Conflicts are easy to miss when only logged with `--verbose`] → Logging at warn level was rejected: the default logger shows warnings, and the extension tree is built on every c8y invocation, so every command would print the message, and the user usually cannot fix the extension. Extension authors can check with `--verbose`.
- [Startup cost] → The tree builder adds one map insert per command. No extra file I/O.

## Migration Plan

No user migration. Ship as its own PR. Rollback is a revert of the PR.

## Open Questions

- Resolved during implementation: conflicts are logged at info level (shown with `--verbose`). Warn level is shown by default, so it would print on every c8y command.
