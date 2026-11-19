# go-c8y v2 integration — input pipeline, service calls, output streaming

Status: prototype on `exp-go-c8y-v2-output-filters`. Working commands:
`devices list2`, `devices create2`, `devices update2`, `template execute2`.

## Model

The v1 pairing `requestiterator.RequestIterator` + `worker.Worker` (one raw
HTTP request materialised per input item) is replaced for v2-based commands by
`pkg/c8ystream`. The front door is the **Runner**. A command configures its
input driver and (optionally) compound builders, then provides one **Build**
callback that, per input item, fills the typed arguments of its service
function and returns the **Call**:

```go
r, err := c8ystream.NewRunner(cmd, factory)   // cfg, pipeline annotations, input cursor
r.InputFlag("id", c8yfetcher.WithDeviceByNameFirstMatch(...))  // id drives iteration
r.Body(flags.WithDataFlagValue(), ...)                          // body builder
client, err := r.Client()

return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
    body, err := in.Body()
    if err != nil { return nil, err }
    id := in.String("id")
    return func(ctx context.Context) output.Seq {
        return c8ystream.FromResult(client.Devices.Update(ctx, id, body))
    }, nil
})
```

The command does **not** classify flags as query vs body vs path. With typed
service calls, go-c8y's options struct decides serialisation (its `url` tags
turn fields into query params; the body is sent as-is). So the command just
reads resolved values and assigns struct fields:

```go
return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
    opt := devices.ListOptions{Query: must(in.Query())}
    opt.WithChildren = in.Bool("withChildren")
    opt.PaginationOptions = ...
    return func(ctx) output.Seq { return output.FromIterator(client.Devices.ListAll(ctx, opt).Items()) }, nil
})
```

```
input driver (pipe line OR primary flag slice)        [reuses v1 iterators]
        │  one item at a time, serial
        ▼
Build(Resolver) ─── command fills the options struct from in.String/Bool/Query/Body
        ▼            → returns a Call closure capturing the resolved values
Stream(ctx, workers, calls) ─── executes the Calls   (worker pool when workers > 1)
        ▼  flat stream of jsondoc.JSONDoc
ErrorCollector → --filter → --outputTemplate → --select   [v2 output stages]
        ▼
Renderer: --output (json/csv/csvheader/tsv), teed to --outputFile   [v2 encoders]
```

- **Resolution is serial, execution is concurrent.** Build runs on the serial
  input-draining goroutine, so `in.*` resolution (and the group/name→id
  lookups inside it) never races; the returned Call closure captures the
  resolved values and runs on the worker pool. The command can't accidentally
  read per-item state concurrently because the Resolver isn't reachable from
  the Call.
- **One input driver.** `Input()` reads piped stdin (one item per line) or runs
  once with no pipe. `InputFlag(name, …)` makes a flag the driver: piped items
  feed it, or its own slice/file values drive iteration (`--id 1,2,3`). The
  driver ends via `io.EOF`, capped at `--maxJobs`.
- **Compound fields** that don't map 1:1 from a flag — the Cumulocity `q`
  expression and the templated body — are configured once (`r.Query(...)`,
  `r.Body(...)`) and evaluated per item (`in.Query()`, `in.Body()`). Everything
  else is a direct `in.String`/`in.Bool`/`in.StringSlice` read in Build.
- `in.String(flag)` resolves `-`/`-.path` input references against the current
  item (see below). `in.Query()` returns the expression decoded once, because
  the flag formatters pre-encode for the v1 raw-query path while the v2 client
  encodes query parameters itself.
- `Stream` with `workers <= 1` is fully pull-based and lazy: `output.Head(n)`
  stops pagination *and* input consumption. With `workers > 1` calls run on a
  pool; documents merge in arrival order and stopping the consumer cancels
  in-flight calls via context.

### Dumb getters — push resolution into go-c8y

The flag getters should be dumb: read a value from a flag or the piped item,
nothing more. The Cumulocity-specific knowledge (filter syntax, name→id
lookups) belongs in go-c8y, which already has it:

- **Query building** — `model.InventoryQuery` (`pkg/c8y/api/model/inventoryQuery.go`)
  is a fluent builder: `.AddFilterEqStr("name", v).HasFragment(f).ByGroupID(id).AddOrderBy(o).Build()`
  → `$filter=(...) $orderby=...`. `list2` now builds its query this way from
  dumb `in.String(...)` reads — no `WithCumulocityQuery`, no per-flag
  `WithStringValue("name","name","(name eq '%s')")` carrying the c8y syntax.
- **Name resolution** — `managedobjects` has a smart resolver: a plain string is
  a literal id (pass-through), `name:foo` / `ext:type:val` / `query:...` are
  looked up. `Devices.Get/Update/Delete` call `ResolveID` internally. So
  `update2` can pass `managedobjects.DeviceRef(in.String("id"))` straight to
  `Update` and let go-c8y resolve — replacing `c8yfetcher.WithDeviceByNameFirstMatch`.
  (Iteration binding — pipe/slice over `--id 1,2,3` — stays CLI-side; go-c8y
  doesn't read stdin. Only the *resolution* moves.)

What's left CLI-side after this: dumb value reads (`in.String/Bool/Time`), the
jsonnet body builder, and a few transforms with no go-c8y equivalent yet —
relative-time encoding (`in.Time`, "-7d" → ISO) and the gaps below.

**Device-group by name — done.** go-c8y now has a `devicegroups` service
(`pkg/c8y/api/devicegroups/api.go`): a device group is a managed object with
`c8y_IsDeviceGroup`, so the service wraps managed-objects and scopes reads and
name resolution to that fragment. `Service.ResolveID(ctx, ref, meta)` turns
`name:<n>` into a group id (group-scoped lookup `has(c8y_IsDeviceGroup) and
(name eq '<n>')`), passing plain ids through; `ByName`/`ByID`/`GroupRef`,
`List`/`ListAll`/`Get`/`Update`/`Delete`/`Create` round it out. `list2 --group`
now resolves a name (all-digit value → id) via `client.DeviceGroups.ResolveID`.

Two integration notes this surfaced:
- The CLI must build against the local go-c8y-v2 to see the new service: added
  `replace github.com/reubenmiller/go-c8y/v2 => ../go-c8y-v2` to the CLI
  go.mod. (This also pulls in the earlier local v2 fixes — the dry-run
  transport guard in `pkg/c8ystream/transport.go` is now redundant but
  harmless; revisit when the module is next published.)
- **Resolution must bypass dry-run.** A name→id lookup is a real API call even
  under `--dry` (so the dry-run shows the resolved request, as in v1).
  `ctxhelpers.ResolutionContext` only strips *deferred* execution, not dry-run,
  so the CLI uses `Resolver.ResolveContext()` (run ctx with dry-run + mocks
  disabled) for lookups.

go-c8y gaps still open:
- **queryTemplate / custom query params** (`cfg.GetQueryParameters()`) — no
  builder hook yet; needs the per-call request-modifier seam.

### Why not Query/Body/Param "sources"

An earlier iteration grouped bindings as `Query(...)`, `Body(...)`,
`Param(...)` sources — a holdover from v1, where a value's HTTP location
(query string vs body vs path) genuinely mattered because the CLI built the
raw request. With typed service calls that distinction belongs to go-c8y, not
the command: a command just links values to its options struct. The Resolver
model drops the grouping — simple fields are direct `in.*` reads, and only the
irreducibly-compound `q` expression and body remain as named builders.

## Input references (`-` / `-.path`) — one item, many options

The v1 model can only bind the pipe to **one** flag (the single pipeline
target marked with `-`/`-.`). The `Input()` source lifts that limit while
keeping the same `-`/`-.path` sentinels: it is the sole stdin reader,
publishing each piped item to a shared cursor (`iterator.InputCursor` on
`RequestInputIterators`). Any string-flag value of the form `-` (whole item)
or `-.path` (a gjson path) then resolves from that cursor per iteration, so
several flags read the same object:

```
echo '{"name":"foo","type":"bar"}' | c8y devices list2 --type -.name --name -.name
#  → $filter=(name eq 'foo') and (type eq 'foo')   (both from the one piped object)
echo "name eq 'x'" | c8y devices list2 --query -    # plain text: whole line into the query
```

`-` / `-.` are the existing `flags.FlagReadFromPipeText` /
`flags.FlagReadFromPipeJSON` sentinels — the new model generalises them from
"the one pipe target" to "any flag, all sharing one item" rather than
introducing new syntax.

Mechanism: `flags.WithStringValue` returns an `iterator.InputRefIterator`
(instead of a static string) when the value is an input reference and a cursor
is active. `CumulocityQueryIterator` already re-evaluates iterator-valued
filter parts per item, so the existing query builders gain multi-bind for
free — the command only adds `c8ystream.Input()` to its `Bind` call.

Scope and rules:
- Works for **query/path string options** (`WithStringValue`, and so
  everything inside `WithCumulocityQuery`). **Body** flags keep using the
  richer jsonnet `input.value.x` mechanism the mapbuilder already provides —
  `-.path` is not wired through the body builder.
- `Input()` is **mutually exclusive** with the legacy single-flag binding: it
  neutralises the command's `PipeOptions` (name cleared, disabled) so nothing
  competes for stdin. The pinned getters dereference `PipeOptions` under only
  an `inputIterators != nil` guard, so it is neutralised in place, not nilled.
- Non-JSON input is fine: `-` yields the whole raw line; `-.path` on a
  non-object yields empty (the option is simply unset).
- No pipe → the command runs once with `-` references unset.

Not yet input-ref aware (follow-ups): `WithPathParameters`/`c8yfetcher` id
resolution (so `update2` stays on the legacy single-pipe model), and the
non-string getters (bool/timestamp). The cursor + interception pattern extends
to them the same way.

## Per-command interactions

| Command | Driver + builders | Build → Call |
|---|---|---|
| `devices list2` | `Input()` + `Query(WithCumulocityQuery(...))` | `ListOptions{Query: in.Query(), …bools}` → `Devices.ListAll` → `FromIterator` |
| `devices create2` | `Input()` + `Body(...)` | `Devices.Create(ctx, in.Body())` → `FromResult` |
| `devices update2` | `InputFlag("id", fetcher)` + `Body(...)` | `Devices.Update(ctx, in.String("id"), in.Body())` → `FromResult` |
| `template execute2` | `Input()` + `Body(...)` | none — `FromDocs(jsondoc.New(formatOutput(in.Body())))` |

For codegen: every command is the same skeleton — `NewRunner`, one
`Input()`/`InputFlag()`, optional `Query(...)`/`Body(...)`, then a `Build` that
assigns resolved values to the service options struct and returns the call.
The generator emits the field assignments (it knows the spec) and the one-line
service call; the query/body builders are emitted only when the command has
those parameter kinds.

Notes:

- **id resolution** keeps the path-parameter + fetcher machinery (including
  name → id lookup) but the template is just `"{id}"`, so the resolved value
  feeds the typed call instead of a URL path.
- **Bodies**: `Create/Update(ctx, body any)` accept `json.RawMessage`, so the
  mapbuilder/jsonnet output passes through untouched.
- **MaxItems** (list): `--includeAll` → 0 (unlimited), `--totalPages N` →
  `N × pageSize`, default → one page (`pageSize`) to match v1.
- **Cross-cutting request options** travel via context: `apiv2.WithDryRun`,
  `c8ystream.WithProcessingMode` (reads `--processingMode`).
- **Dry run** disables the v2 mock responses (`WithMockResponses(ctx, false)`)
  so it only reports the calls that would be made, like v1. The pinned v2
  version then sends the real request (transport bug, fixed upstream in the
  dry-run middleware: dry run without mocks must return an empty 204, never
  send) — `c8ystream.newDryRunTransport` guards this at the CLI's base
  transport until the module is bumped, after which it can be removed.
- **--output**: json → NDJSON, csv/csvheader/tsv → the v2 CSV/TSV encoders
  (columns derived from the first document, composing with --select); table
  falls back to NDJSON until the console/dataview renderer is ported.
- **--verbose/--debug** enable the v2 client's request/response debug
  logging. Note the v2 library logs via slog (plain format) while the CLI
  logger is styled — bridging slog into the CLI logger is an open task.
- **Client construction**: `Runner.Client()` builds the v2 client from the
  *legacy factory client*, which owns credential resolution (encrypted
  session decryption, passphrase prompting, token refresh) — building it from
  raw config values fails on encrypted sessions ("decryption failed"). TODO:
  move onto `cmdutil.Factory` (`ClientV2`) and carry proxy/TLS/timeout.

## Teeing

Three hook points, all composable:

1. **Sink**: `c8ystream.Multi(...)` fans documents to several renderers with
   different encodings — used by `--outputFile` (NDJSON on terminal, JSON
   array in file).
2. **Stage**: `c8ystream.Tap(fn)` passes documents through while handing them
   to fn (e.g. backup of processed items mid-pipeline).
3. **Request level** (not yet ported): activity log belongs in a resty
   response middleware on the v2 client, not in the output pipeline.

## Error semantics

`ErrorCollector` (first stage) records item-level errors and lets the stream
continue (one bad input doesn't kill the batch); `Max` reproduces
`--abortOnErrors`. After `Render`, `collector.Err()` is the command result so
the exit code reflects partial failure.

## go-c8y v2 gaps found

- **Arbitrary query params/headers per typed call** (the request-modifier
  seam): `cfg.GetQueryParameters()` / custom headers have nowhere to go in
  `ListOptions`. Suggest a context-carried request mutator or variadic
  `core.RequestOption` on service methods. Until then those options are
  silently unsupported in `*2` commands.
- **`PaginationOptions.CurrentPage` is ignored by `pagination.Paginate`**
  (iterator always starts at page 1), so `--currentPage` has no effect on
  `list2`.
- **Processing mode was not applied by the execute path**:
  `ExecuteResponseOnly` used resty's `SetContext`, bypassing
  `TryRequest.ApplyProcessingModeFromContext`. Fixed in go-c8y-v2
  (`pkg/c8y/api/core/execute.go`); the CLI picks it up on the next module
  bump. The deferred-execution capture branches still bypass it (cosmetic —
  the executor applies it on the real send).
- **Dry-run logging** prints method/url/headers but not the request body;
  v1 dry-run rendered the body. Worth adding before replacing v1 commands.

## Gaps surfaced by promoting devices CRUD

`devices list/get/create/update/delete` now run on the v2 streamer (the v1
`.auto.go` were replaced; `list2`/`create2`/`update2` removed). The id flag
drops `c8yfetcher` entirely: `InputFlag("id")` binds iteration only (pipe
extracts id/deviceId/... aliases, or the `--id 1,2,3` slice), and
`c8ystream.NameOrID(in.String("id"))` → `Devices.Get/Update/Delete` lets go-c8y
resolve (all-digit = id, else `name:`). Verified: id slice → N calls, piped
device object → id extracted, `--id <name>` → real `(name eq '…')` lookup even
under `--dry` (the `ResolutionContext` fix).

What the v1 commands did that the v2 ones don't yet — the gap list:

- **Custom headers** (`--header`) — not sent. Needs a per-call request seam.
- **Custom query params** (`cfg.GetQueryParameters()` / `--customQueryParam`)
  — not applied. Same seam.
- **`--queryTemplate`** (list) — ignored; only `--query -` (raw line as filter)
  works. The `%s`-into-template formatting has no `InventoryQuery` equivalent.
- **`--withGroups`** (list, get) — no field in `managedobjects.GetOptions` /
  devices `ListOptions`; silently dropped.
- **Implicit pipe → designated flag** (create) — v1 `echo names | create` bound
  the pipe to `name`; the v2 `Input()` model requires explicit `--name -`.
- Still unported (deliberate): `--outputFileRaw`, activity log, confirmation
  prompts, `--delay`/`--delayBefore`.

The first three collapse into one go-c8y addition — a **per-call request
modifier** (context-carried `func(*resty.Request)` or a variadic
`core.RequestOption` on service methods) so the CLI can inject arbitrary
headers/query params and a templated query. That's the next seam to build.

## Progress

`--progress` is wired in `pkg/c8ystream/progress.go`: a single mpb spinner on
stderr, ticked by a `Tap` stage as items reach the renderer (the stages run on
the one consuming goroutine, so no locking). It uses mpb directly — the
`pkg/progressbar` wrapper is worker-pool shaped and its `Wait()` is
self-recursive. Counts output items, not requests; per-page/per-request
counting would need a hook in the pagination loop.

## Cross-cutting command concerns (ported)

The seven concerns that previously lived in the v1 worker + request handler are
now on the v2 streamer. They split by *when* they need the request/response:

- **Submission side (pre-execution, serialized)** — `pkg/c8ystream/submission.go`.
  A `submission` (built once per command from the session config, carried in
  context) handles **confirmation prompts** and the **`--delay`/`--delayBefore`**
  timing. `Submit`/`SubmitStatus` wrap a single mutating call: when confirmation
  is active they run it under `WithDeferredExecution` to capture the prepared
  `*http.Request` (the v2 dry-run transport returns an empty 204, so the
  mutating request is never sent before the prompt), prompt serially (a mutex
  shares the "yes/no to all" answers and the prompt counter across items), then
  `Execute` for real. Delays wrap every job in the producer. `confirmActive()`
  is `cfg.ShouldConfirm()`, so CI/`--force`/`--dry` keep the plain `FromResult`
  fast path. The mutating commands call `Submit`/`SubmitStatus`; `get`/`list`
  stay on `FromResult`/`ListCall` (GETs are never confirmed).
- **Transport level** — `pkg/c8ystream/activitylog.go`. The **activity log** is a
  resty response middleware on the v2 client (added in `Runner.Client`), so it
  records every request — pagination pages and name→id lookups included —
  uniformly. Dry-run/deferred-prepare responses (`X-Dry-Run` header) are skipped.
- **Output pipeline (post-execution)** — `--outputFileRaw`
  (`outputfileraw.go`) is a `Tap` stage placed *before* filter/select so the
  file holds the unshaped response, with `{id}/{name}/{type}/{owner}` filename
  templating. **table/csv views** (`view.go`) is a shaping stage that restricts
  documents to the detected (`--view auto`, by fragments) or named view's
  columns via `shape.Select`, applied to all formats like v1; it is a no-op for
  `--select`/`--raw`/`--view off`/the non-TTY default, so it never changes the
  current behaviour of the test suite.
- **Error handling** — `pkg/c8ystream/errors.go`. **Silent status codes** needed
  the v2 `core.Error` turned into a `cmderrors.CommandError` carrying the HTTP
  status + mapped exit code (`toServerError`); `Run` then routes the collected
  errors through `Factory.CheckPostCommandError`, so `--silentStatusCodes`,
  `--silentExit`, the exit code (4 for 404, 104 for a multi-error batch) and
  error printing match the v1 worker path. Previously every v2 error surfaced as
  an "unexpected error" (exit 100, status-blind).
- **Context** — `WithProcessingMode` (`client.go`) now falls back to the
  **session-default processing mode** (`settings.defaults.processingMode`) when
  `--processingMode` is unset.

Number formatting: `numbers.RawNumber` (the `--view ... numberFormat none`
formatter) preserves the source token (`0.000000000001`, not `1e-12`); a stray
`json.Marshal` round-trip introduced by the streamer commit was reverted.

Validation: an out-of-process fake Cumulocity server now exists at
`go-c8y-v2/cmd/c8y-fakeserver` (see that repo) for offline end-to-end CLI tests.
