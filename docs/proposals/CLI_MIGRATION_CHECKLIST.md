# CLI service migration — handoff checklist

Goal: migrate go-c8y-cli commands from spec-generated `*.auto.go` to hand-written
`c8ystream.Runner` commands against the **go-c8y v2 SDK**, carrying the
annotations that let the tree project to PowerShell/docs.

Read first: `go-c8y-v2/docs/proposals/CLI_CODEGEN_INVERSION.md` (the why + architecture).

Branch: `exp-go-c8y-v2-output-filters`. Both repos are siblings under
`go-c8y-nextgen/` (`go-c8y-cli`, `go-c8y-v2`). go-c8y-cli's `go.mod` has a local
`replace github.com/reubenmiller/go-c8y/v2 => ../go-c8y-v2` — **do not commit
go.mod** (local-only; not in HEAD).

## Reference implementations — copy these, don't start from scratch

| Pattern | Copy from |
|---|---|
| Source-bearing CRUD (source.id) | `pkg/cmd/events/{list,get,create,update,delete}/*.go` |
| Multi-ref + enum validate-set + ResolveBodyRef | `pkg/cmd/operations/{list,get,create,update}/*.go` |
| Enum-slice list filters + validate-sets | `pkg/cmd/alarms/{list,create,update}/*.go` |
| No-resolution plain CRUD | `pkg/cmd/auditrecords/{list,get,create}/*.go` |
| Inventory-query (q-builder) list | `pkg/cmd/devices/list/list.go` |
| Bridge primitives | `pkg/c8ystream/resolver.go` (`TimeValue`, `ResolveSourceID`, `ResolveBodyRef`, `NameOrID`) |
| Stream helpers | `pkg/c8ystream/{list.go (ListCall), c8ystream.go (FromResult/FromStatus), submission.go (Submit/SubmitStatus)}` |

## Done so far (do not redo)

- Infra: annotation vocabulary (`pkg/flags`, `pkg/completion`), `internal/clibuild`,
  `internal/clisurface`, `cmd/gen-surface`, tree-walking `gen-powershell --from-tree`,
  the `c8ystream` bridge.
- Migrated + e2e-verified: **devices, events, alarms, measurements, operations, auditrecords,
  retentionrules, tenants, bulkoperations**.
- SDK: added `Measurements.Get` (go-c8y-v2 commit `09cdc2a`); added bulkoperations
  `ListOptions` filters withDeleted/dateFrom/dateTo/generalStatus (go-c8y-v2 commit `e47d2d8`).

## Per-service recipe

### 0. Triage the service (decide clean vs needs-design)
```bash
SVC=tenants   # cli group name
# SDK surface:
grep -nE "^func \(s \*Service\) (List|ListAll|Get|Create|Update|Delete)\b" ../go-c8y-v2/pkg/c8y/api/$SVC/api.go
grep -rnE "type .*Iterator = pagination.Iterator|DeviceResolver \*" ../go-c8y-v2/pkg/c8y/api/$SVC/
sed -n '1,80p' ../go-c8y-v2/pkg/c8y/api/$SVC/zz_generated_options.go   # ListOptions fields
# CLI surface:
grep -nE "cmd[A-Z][a-zA-Z]*\.New" pkg/cmd/$SVC/$SVC.auto.go
grep -nE 'Flags\(\)\.|WithExtendedPipelineSupport|WithValidateSet' pkg/cmd/$SVC/list/list.auto.go
# PS names + accept media-types (match the existing cmdlets exactly):
for f in tools/PSc8y/Public/Get-Tenant*.ps1; do grep -H "Type = " "$f"; done
```
- **Clean** if: every CLI list flag maps to a typed `ListOptions` field (or the
  `q`-builder), and create/update bodies are typed flags + `--data`/`--template`.
- **Needs design** if: list flags have no typed field and no clean `q` mapping
  (gaps), or the command shape isn't collection-CRUD (e.g. identity).
- If the SDK service is **missing `Get`/`Delete`** etc., add it to go-c8y-v2
  (mirror `events`/`alarms`; commit separately in go-c8y-v2) — see `Measurements.Get`.

### 1. Write each command (`pkg/cmd/$SVC/$cmd/$cmd.go`, package + constructor name unchanged)
Keep `Use`, `Short`, `Long`, examples, `PreRunE` (Create/Update/DeleteModeEnabled),
flags. Then:
- **Annotations** (in `flags.WithOptions`):
  - `flags.WithPowershellName("Get-Foo")` — match the existing cmdlet name exactly
  - `flags.WithOutputType("<accept media-type>", "<item media-type or empty>")` — from the spec/old cmdlet `Type`/`ItemType`
  - `flags.WithCollectionProperty("foos")` — on `list` only
  - `flags.WithExtendedPipelineSupport(<pipeFlag>, <property>, <required>, <aliases...>)` — the pipe target
  - `completion.WithValidateSet("status", "A", "B")` — enums (auto-projects to PS `[ValidateSet]`)
- **RunE**:
  - `r, _ := c8ystream.NewRunner(cmd, n.factory)`
  - iteration: `r.InputFlag("id"|"device")` for id/filter-driven; `r.Input()` for create
  - `client, _ := r.Client()`
  - list → fill typed `ListOptions`, set `PaginationOptions` (copy from a reference), `return c8ystream.ListCall(rawOutput, opt, client.Foo.ListAll)`
  - get → `return c8ystream.FromResult(client.Foo.Get(ctx, id))`
  - create → `r.Body(...)`, then `in.ResolveBodyRef(body, "<path>", resolveFn)` per device ref, then `c8ystream.Submit(ctx, func... client.Foo.CreateRaw(ctx, body))`
  - update → `r.Body(...)` then `c8ystream.Submit(ctx, func... client.Foo.Update(ctx, id, body))`
  - delete → `c8ystream.SubmitStatus(ctx, func... client.Foo.Delete(ctx, id))`

### 2. Delete the replaced `.auto.go`
```bash
git rm pkg/cmd/$SVC/{list,get,create,update,delete}/*.auto.go   # only the ones you replaced
```
(The `$SVC.auto.go` wiring calls the same `New*Cmd` names — it keeps compiling.)

### 3. Verify
```bash
gofmt -l pkg/cmd/$SVC/
go build ./pkg/cmd/$SVC/... ./cmd/c8y          # must be clean
# tree projection — confirm cmdlet names, [ValidateSet], common-param sets:
go run ./cmd/gen-powershell --from-tree --command $SVC --module /tmp/ps$SVC && ls /tmp/ps$SVC/Public
# fakeserver e2e (see below) — at least create-by-name + list + one round-trip
go test ./pkg/c8ystream/ ./internal/clisurface/ ./internal/codegen/powershell/
```

### 4. Commit (one per service; SDK changes commit separately in go-c8y-v2)
```
feat($SVC): migrate <commands> to v2 c8ystream commands
...body: what maps to typed options, what resolves, any GAP...

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
```

## Fakeserver e2e (offline, no tenant)
```bash
# build + run server
(cd ../go-c8y-v2 && go build -o /tmp/c8y-fakeserver ./cmd/c8y-fakeserver)
/tmp/c8y-fakeserver --addr 127.0.0.1:8111 &     # seeds tenant t12345, user admin,
                                                #   devices TestDevice001(10004), thin-edge-device-001(10005)
go build -o .bin/c8y ./cmd/c8y
export C8Y_HOST=http://127.0.0.1:8111 C8Y_TENANT=t12345 C8Y_USER=admin C8Y_PASSWORD=admin-pass C8Y_SETTINGS_CI=true
# always append </dev/null to a non-piped c8y command (it blocks on stdin in non-TTY)
.bin/c8y $SVC create --device TestDevice001 ... --output json </dev/null   # by NAME → source.id must resolve to 10004
.bin/c8y $SVC list ... --output json </dev/null
pkill -f /tmp/c8y-fakeserver
```
Caveat: fakeserver CQL/`q` filtering is a subset — some `--query` results may differ.

## Gotchas (these bit me)
- **Body flags must be scalar `String`, not `StringSlice`.** A body getter
  (`WithStringValue`/`WithOverrideValue`) on a `StringSlice` flag errors
  "trying to get string value of flag of type stringSlice". On `create`/`update`,
  declare `device`/`agent`/etc. as `String`. (List filters driven by `InputFlag`
  can stay `StringSlice`.)
- **Source/device resolution lives in the body path**, not source.id always:
  events/alarms/measurements use `source.id`; operations use `deviceId`+`agentId`.
  Use `in.ResolveBodyRef(body, "<path>", resolveFn)` once per ref. `resolveFn` =
  `client.<Svc>.DeviceResolver.ResolveID(ctx, managedobjects.DeviceRef(ref), nil)`
  (or `client.DeviceGroups.ResolveID` for group refs).
- **`in.Body()` returns json.RawMessage**; `ResolveBodyRef` returns json.RawMessage
  (NOT []byte — passing []byte to CreateRaw base64-encodes it).
- **Typed time fields**: use `in.TimeValue(flag)` (→ time.Time), not `in.Time` (string).
- **Enum slices**: convert CLI strings, e.g. `[]model.AlarmSeverity{model.AlarmSeverity(s)}`,
  `types.OperationStatus(s)`. Validation values come from `completion.WithValidateSet`.
- **`*Iterator` must be a type alias** (`= pagination.Iterator[T]`) for `ListCall` to
  accept `client.Foo.ListAll`. They are, for the migrated services.
- **No delete-by-id** for some resources (alarms cleared not deleted; operations) —
  leave `deleteCollection` on spec.
- **Known SDK-gap pattern**: list flags the typed `ListOptions` can't express
  (measurements `--csvFormat/--excelFormat/--unit`; inventory `--owner/--onlyRoots`).
  Carry the supported ones, drop+document the rest, OR add a request-modifier seam
  to go-c8y first.

## Remaining worklist

> **Full per-command inventory with effort + SDK-call + priority columns:
> [`CLI_MIGRATION_TRACKER.md`](CLI_MIGRATION_TRACKER.md)** (auto-generated, one row per command,
> 185 commands). The lists below are a curated subset; the tracker is the source of truth.

### Clean (mechanical — follow the recipe)
- [x] **bulkoperations** — done (commit `b68f46be`). `group`→`groupId` via `client.DeviceGroups.ResolveID` + `ResolveBodyRef`; `group` flag made scalar String. Needed an SDK change (list filters: go-c8y-v2 `e47d2d8`). listoperations stays spec.
- [x] **tenants** — done (commit `a92f0740`). Plain CRUD; `id`/`parent` default to current tenant via `GetTenant()`; `--name`→`company`. Delete passes `tenants.DeleteOptions{}`. enable/disable/tfa/applications/assert/listreferences stay spec.
- [x] **retentionrules** — done (commit `1232b920`). Plain CRUD, no refs; `dataType` ValidateSet; no SDK change.
- [ ] ~~**smartresttemplates**~~ — NOT a migration: no existing `pkg/cmd/smartresttemplates` group, and the SDK uses Options structs (`CreateOptions`, Get-by-name, no Update). Would be net-new commands.
- [ ] ~~**trustedcertificates**~~ — NOT a migration: no existing `pkg/cmd/trustedcertificates` group; SDK uses Options structs (Create/Get/Update/Delete all take `*Options`). Net-new commands.

### Needs design decision (do not mechanically convert — decide first)
- [ ] **inventory / managedObjects** — `list` mixes typed fields (`text`/`ids`/`query`) with q-builder filters (`owner`) and has gaps (`onlyRoots`, `childAdditionId`). Decide q-vs-typed; model on `devices/list`. get/create/update/delete are clean (managedobjects resolves names in Get/Update/Delete; create has no source — like devices create minus the `c8y_IsDevice` fragment). PS: `*-ManagedObject(Collection)`.
- [ ] **identity** — external IDs are not collection-CRUD (get/create/delete by `type`+`value`). Different command shape.
- [ ] **applications / appversions** — list variants (by name/tenant/user); special bodies.
- [ ] **users / usergroups / userroles / inventoryroles / devicepermissions** — nested structures, references.
- [ ] **microservices / repository (firmware, software, configuration, deviceprofiles) / binaries** — binary upload + repository-backed; complex.

### After enough coverage (end-state — separate work)
- [ ] Point `gen-tests` at `clisurface` (tree-based test gen) so flipping doesn't lose the 281 Pester tests.
- [ ] Flip `task build-powershell` to `--from-tree`; retire the spec PowerShell generator (`powershell.Generate(specDir)` + `GenerateSpecFile`).
- [ ] Teach `gen-cli` (`internal/codegen`) to skip migrated commands — its golden (`TestGoldenSpecOutput`) reports "missing .auto.go" for every migrated command (cosmetic, accumulating).
- [ ] Add the per-call request-modifier seam to go-c8y (unblocks `--header`/`--customQueryParam`/`--csvFormat`/`--unit` gaps).
- [ ] Decide on `go.mod`: published go-c8y/v2 version bump vs keeping the local replace.
