## 1. Tree builder

- [x] 1.1 Create `pkg/extensions/cmdtree` with `Node`, `Leaf`, `SourceKind` (`SourceAPI`, `SourceScript`), `AddGroup`, `AddLeaf`, and `ToCobra`
- [x] 1.2 Implement leaf conflict precedence (API over script) with a warning callback that receives path, kept origin, and dropped origin
- [x] 1.3 Implement group-over-leaf conflict handling with a warning
- [x] 1.4 Implement description merging (explicit description replaces generated placeholder; first explicit description wins)
- [x] 1.5 Unit tests: #585 merge, duplicate spec groups, deep nesting (4+ levels), leaf conflict, group-over-leaf conflict, description merge, deterministic output

## 2. API spec source

- [x] 2.1 Split `cmdparser.ParseCommand` into a spec parser that returns group metadata and per-command builders; keep `ParseCommand` as a wrapper with unchanged behavior
- [x] 2.2 Test that `ParseCommand` produces the same commands as before for every file in `tests/testdata/extensions/c8y-kitchensink/api`

## 3. Wire into root command

- [x] 3.1 Rewrite `ConvertToCobraCommands` in `pkg/cmd/root/root.go` to add API spec groups and leaves, then `commands/` leaves, to one tree per extension, then attach `tree.ToCobra()` to the root
- [x] 3.2 Keep existing script execution behavior (flag parsing disabled except during tab completion, `FParseErrWhitelist`, executor callback)
- [x] 3.3 Only attach the extension root when it has at least one child

## 4. Fixtures and integration tests

- [x] 4.1 Add a script under an API-defined group to `tests/testdata/extensions/c8y-kitchensink` (for example `commands/<api-group>/script-cmd`) and a deeply nested script
- [x] 4.2 Add a test in `tests/manual/extensions` (or `tests/auto` if no tenant is needed) that runs both commands from #585 and the nested command
- [x] 4.3 Run `go test ./pkg/...` and the extension test suite

## 5. Docs

- [x] 5.1 Document in the extension concepts docs (`docs/go-c8y-cli/docs/concepts/extensions`) that API and script commands can share groups, the nesting rule, and the conflict precedence
- [x] 5.2 Reference #585 in the PR description (CHANGELOG.md is not maintained; release notes come from GitHub releases)

## 6. Nested API command groups (#562)

- [x] 6.1 Split `group.name` on whitespace into the group path, and declare groups for specs without commands (so they can set a parent group's description)
- [x] 6.2 Unit test for nested API groups and parent group descriptions in `pkg/extensions/cmdtree`
- [x] 6.3 Add `api/hierarchy.yaml` and `api/hierarchy_items.yaml` to the kitchensink fixture, update the `cmdparser` golden file, and add commander tests for `c8y kitchensink hierarchy items ...`
- [x] 6.4 Document nested groups in `tools/schema/extensionCommands.json` and in `docs/go-c8y-cli/docs/concepts/extensions/02-api-commands/01-command-groups.md`
