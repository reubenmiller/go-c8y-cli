# go-c8y-cli on go-c8y v2: adoption assessment and target architecture

Status: assessment, 2026-11-19
Scope: `go-c8y-cli` branch `exp-go-c8y-v2-output-filters` (including staged changes) and `go-c8y-v2` at `7f3e104`.
Replaces: `docs/proposals/CLI_MIGRATION_CHECKLIST.md` (now out of date, see section 1).
Progress: see section 8 for what the `v3` branch has implemented.

## Summary

The command-by-command port is essentially finished. No `*.auto.go` files remain, and about 315 commands run on the `pkg/c8ystream` Runner. The remaining risk is **consistency**. The cross-cutting behavior (dry run, confirmation, resolution, pagination, result rendering) is split three ways:

- The SDK handles part of it.
- The `c8ystream` bridge handles part of it.
- Each command makes its own choice from several equivalent helpers.

Every one of those per-command choices can be made wrong, and some already are.

**Recommendation:**

1. **Move every request-level concern into the SDK.** That covers dry-run rendering, request approval (confirmation), reference resolution (including references in bodies) and pagination defaults. Do this **before v2.0.0 is tagged**. Tagging is the only open task in `openspec/changes/prepare-v2-release`, and several of the fixes break the API.
2. **Shrink the CLI command to one shape.** A command declares flags, then maps one input item to one typed SDK call and returns the result. It never chooses how to confirm, resolve, paginate or render.
3. **Enforce that shape with tree-level lint tests**, so drift fails CI instead of slipping through review.

## 1. Where things actually stand

| Item | Checklist says | Actual |
|---|---|---|
| Spec-generated `.auto.go` commands | inventory, identity, applications, users, repository and others remaining | **0 remaining** (`CLI_MIGRATION_TRACKER.md`) |
| Commands on `c8ystream.Runner` | 9 services | ~315 files |
| Commands still on the v1 path (`worker`, `c8yfetcher`, `pkg/request`) | not tracked | ~35 files (listed below) |
| Request-modifier seam (headers, custom query params) | open SDK gap | **Done**: `api.WithExtraHeaders` / `api.WithExtraQueryParams` are applied centrally in `Runner.Run` |
| Regression list from the manual tests | not tracked | `v2-migration-remaining-issues.md`: section A fixed, B1 needs a decision, section C needs a tenant |
| v2 release | not tracked | All tasks done except "Tag v2.0.0" |

**v1 holdouts.** These are commands still on the v1 worker or v1 client. They are the reason `go.mod` still needs `github.com/reubenmiller/go-c8y` v1.

- **Realtime:** `{alarms,events,measurements,operations,inventory}/subscribe`, `realtime/{subscribe,subscribeall}`, `notification2/subscriptions/subscribe`.
- **Wait and assert:** `operations/wait`, `inventory/wait`, `assert/text`, `devices/assert`, `inventory/assert` (which the `count` asserts reuse), `tenants/assert/exists`.
- **Certificate authority:** `devicemanagement/certificate_authority/*`.
- **Device registration:** `deviceregistration/registerBulk/*`, `devices/enroll`.
- **Applications:** `applications/{createhostedapplication,open}`, `microservices/create`, `microservices/serviceuser/*`, `ui/applications/plugins/*`, `ui/plugins/create`.
- **Other:** `inventory/parents/get`, `measurements/createBulk`, `api` (raw requests), `remoteaccess/{connect/*,server}`, `extension`.
- **Completion callbacks:** 202 v2 command files import v1 *only* for their completion callbacks (`func() (*c8y.Client, error)`).

## 2. Consistency findings

These are the reasons the interface is not yet uniform. The counts come from a survey of `pkg/cmd`, and the key items were spot-checked.

### 2.1 Confirmation is opt-in per command, and it has already been missed

The command picks a wrapper from `Submit`, `SubmitStatus`, `SubmitUpload` or `SubmitValue`. Calling `FromResult` or `FromJSONValue` directly skips the prompt silently.

- **Bug:** `tenantoptions/updatebulk/updateBulk.go:108` is gated by `UpdateModeEnabled` but calls `FromJSONValue(client.Tenants.Options.UpdateByCategory(...))`, so it never asks for confirmation.
- **Hang workaround:** `SubmitUpload` exists only because deferred preparation of a multipart body hangs: the `io.Pipe` has no reader. It confirms from a fake `http.NewRequest` stub, so the prompt never shows the real request.
- **Double send:** the confirmation path prepares each request with a dry-run send, then sends it again for real. The dry-run handler fires during preparation, which is why `prepareContext` installs a no-op handler.

### 2.2 Two equivalent resolution styles plus local helpers

- **Lazy:** `c8ystream.NameOrID(...)` passed to an SDK method that resolves internally (148 files).
- **Eager:** `client.X.ResolveID(in.ResolveContext(), ...)` inside Build (105 call sites). `in.ResolveContext()` duplicates the SDK's internal `ResolutionContext`.
- **Body references:**
  - `in.ResolveSourceID` and `in.ResolveBodyRef`, each with a resolver closure written per command.
  - The device concept is resolved through five different services' `DeviceResolver` (Alarms, Events, Measurements, Operations, Notification2).
  - The cause is that `CreateRaw` does not resolve, while the typed `Create(ctx, CreateOptions{Source: DeviceRef})` does (`events/api.go:128-166`).
- **Local helpers:**
  - `groupRef` is duplicated in `devices/list` and `agents/list`.
  - `versionRef`/`patchRef`/`softwareVersionRef` are copied into 7 files.
  - `FirstID` is used in 6 files.
- **CLI-side naming convention:** `NameOrID` (all digits means an id, otherwise `name:`) is a CLI convention every SDK consumer would want. It lives in the CLI.

### 2.3 Pagination is copied by hand into 57 files and has already drifted

- Each list command hand-builds an 8-line `pagination.PaginationOptions{...}` literal, plus `RawOutput`/`MaxItems`/`PaginationStrategy` reads.
- 10 list commands have no `Strategy`: `inventory/{children,assets,additions}/list`, `devicegroups/{children/list,listAssets}`, `devices/{children/list,listChildren,listAssets,statistics/list,services/list}`. This may be intentional for child endpoints. Even so, it is a decision each command repeats instead of a property of the SDK service.
- `FullWalk` (the `--includeAll` dry-run fix) is set only in `devices/list`.
- `applyQueryTemplate` is copied into 17 files.
- **SDK gap:** `ListAll` ignores `CurrentPage`, so `--currentPage` has no effect on lists.

### 2.4 The command chooses how its result is rendered

- **Adapters:** `FromResult` (63), `ListCall` (57), `output.FromIterator` (14), `FromJSONValue` (3), `FromDocs` (2), and 4 hand-written `func(yield ...)` closures. These are crossed with the four `Submit*` variants.
- **Dry-run checks:** six commands check `apiv2.IsDryRun(ctx)` by hand to suppress a zero-value result.
- **Raw output:** `--outputFileRaw` streaming is reimplemented in `binaries/get`, `events/downloadbinary` and `measurements/list`.

### 2.5 Flags for the same concept differ in type and behavior

- **`--device`:** `StringSlice` in 47 places and `String` in 7. The default is `[]string{}` in some places and `[]string{""}` in others.
- **Multi-device fan-out** (`ApplyOverrideSlice`) exists only in alarms/events/measurements/operations create, `bulkoperations/create`, `operations/update` and `deviceregistration/register`. Other `StringSlice --device` creates accept several values but don't fan out: `identity/create`, `remoteaccess/*`, `notification2/subscriptions/create`, `software/versions/install`.
- **Other flags:** `--id` is `StringSlice` in 124 places and `String` in 29. `--group` is `StringSlice` in 14 and `String` in 3.
- **Body boilerplate:** all 94 `r.Body(...)` calls repeat the same three getters (`WithDataFlagValue`, `WithTemplateValue`, `WithTemplateVariablesValue`).

### 2.6 Dry-run output and the client still depend on v1

- **Dry-run output:** `transport.go` renders dry runs through the v1 `RequestHandler.PrintRequestDetails`.
- **Client construction:** `Runner.Client()` builds the v2 client from the v1 client's credentials. It also installs the response cache, a gzip-decompress transport, request logging, the activity log, output-template capture and the success message. It does this **once per command invocation**, so it is a fixed copy of factory logic inside the Runner.
- **Token redaction:** a Bearer token is not redacted in `--dryFormat curl` (remaining-issues C1). This follows from redacting with the v1 client's `Token` field.

### 2.7 SDK-side inconsistencies the CLI papers over

- **`ResolutionContext`** (`internal/contexthelpers/deferred.go:37`) returns `context.Background()` under deferred execution. Any lookup during a confirmation prepare therefore loses `WithAuth`, `WithTenant`, extra headers and processing mode.
- **`WithDryRun(ctx, true)`** silently enables mock responses. The CLI must call `WithMockResponses(false)` *after* it, and the order matters.
- **Source resolution in `ListAll`** runs again on every page (`events/api.go:72-98`; the same applies to alarms, measurements and operations).
- **Method shapes vary:**
  - `Get(ctx, id)` vs `Get(ctx, id, GetOptions)` vs `Get(ctx, GetOptions{ID})`.
  - `Delete` with and without options.
  - Three different `ResolveID` signatures: `(ctx, ref, meta)`, `(ctx, tenant, ref)`, `(ctx, moID, ref)`.
  - `binaries.Get` returns the file, not the metadata.
  - `events.ListV2`.
- **Reference support varies:**
  - events, alarms, measurements and operations take plain ids on Get/Update/Delete.
  - `users.UserRef` has no name resolver.
  - devicegroups/agents/deviceprofiles take typed `XRef`, while managedobjects/applications take a resolvable `string`.
- **No public per-call request hook** for typed service calls. Only `SendRequest` has `PrepareRequest`.
- **No request renderer** (curl/markdown/json/dump) exists in the SDK. Every consumer writes its own.

## 3. Target architecture: who owns what

The rule: **if a behavior depends on the HTTP request or on Cumulocity semantics, the SDK owns it. If it depends on the terminal, flags, stdin or the output format, the CLI owns it.** Each concern has exactly one owner and one code path.

| Concern | Owner | Mechanism |
|---|---|---|
| Building requests, `url` tags, accept types | SDK | Typed service methods (unchanged) |
| Reference resolution (path ids, option fields, **raw body fields**) | SDK | Uniform `Ref` handling, see S5 |
| CLI reference convention (digits means id, otherwise name) | SDK | `source.ParseRef(string)`, replacing `c8ystream.NameOrID` |
| Dry run | SDK | `DryRunTransport` (unchanged) with no implicit mocks, see S3 |
| Dry-run **rendering** (json/curl/markdown/dump plus redaction) | SDK | New `api/inspect` formatters, see S4 |
| Confirmation **decision point** | SDK | Request approver hook, see S1 |
| Confirmation **prompt UI**, yes/no to all, `--force` | CLI | One approver installed by `Runner.Run` |
| Pagination strategy, start page, full-walk | SDK | Per-service defaults; `CurrentPage` honored, see S7 |
| `--pageSize/--includeAll/--totalPages/--currentPage/--raw` mapped to options | CLI, centrally | `r.Pagination()`, see C2 |
| Result to documents (204, dry run, deferred, binary, plain values) | SDK | `output.Source` implemented by `op.Result` and `Iterator`, see S8 |
| Headers, custom query params, processing mode, no-accept | SDK context options (done) | Applied once in `Runner.Run` (done) |
| stdin and `-.path` input, fan-out, `--maxJobs`, workers, delays | CLI | Runner (unchanged) |
| Body templating (jsonnet, `--data`, `--template`) | CLI | `r.Body(...)` with the common getters defaulted |
| Filter/select/template/view/encoders | CLI on top of SDK `output` | Runner stages (unchanged) |
| Session, credentials, cache, activity log | CLI | `Factory.ClientV2()`, built once, see C6 |

### The target command shape

Every command has the same RunE. Mutating, read and list commands differ only in which SDK method they call:

```go
func (n *CreateCmd) RunE(cmd *cobra.Command, _ []string) error {
    r, client, err := c8ystream.Setup(cmd, n.factory)   // runner + client + common getters
    if err != nil {
        return err
    }
    r.Input()
    r.Body(flags.WithRelativeTimestamp("time", "time"), flags.WithStringValue("type", "type"), ...)

    return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
        body, err := in.Body()   // device refs already normalized to "name:x" or an id by the ref flag kind
        return func(ctx context.Context) any {
            return client.Events.CreateRaw(ctx, body)   // SDK resolves source.id, approval and dry run happen in the transport
        }, err
    })
}

// list:
return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
    opt := events.ListOptions{Source: in.Ref("device"), Type: in.String("type"),
        PaginationOptions: r.Pagination()}
    return func(ctx context.Context) any { return client.Events.ListAll(ctx, opt) }, nil
})
```

This shape removes `Submit*`, `ListCall`, `FromResult`/`FromStatus`/`FromJSONValue`, `ResolveContext`, `ResolveSourceID`/`ResolveBodyRef`, the resolver closures, the pagination literals and the `IsDryRun` checks from every command. A command can no longer skip confirmation or dry run, because neither is something a command does.

## 4. Required SDK changes (go-c8y v2), ordered by when they must land

**Before tagging v2.0.0 (these break the API or change behavior):**

- **S1. Request approval hook.**
  - **API:** `api.WithRequestApprover(ctx, func(*http.Request) (approve bool, err error))`. `DryRunTransport` (or a layer just inside it) calls the approver before sending any request whose method is not GET or HEAD. It runs after resolution, so the approver sees the final URL and body.
  - **Uploads:** for streaming bodies the approver gets the request before the body is read. If the approver declines, the transport closes the body pipe, so uploads need no special case.
  - **Exemptions:** lookups made under `ResolutionContext` are exempt.
  - **Requests with several sends:** for operations that send several requests (upload then update, get-or-create), add `api.WithApprovalScope(ctx)`. The first decision in a scope covers the rest of that scope's requests, and the CLI opens one scope per input item.
  - **Effect:** this replaces the prepare-then-execute double send and all four `Submit*` variants. Deferred execution stays for library users who want to inspect or schedule requests.
- **S2. Fix `ResolutionContext`.** It must derive from `ctx` by clearing the dry-run, deferred, mock and approver flags, not by starting from `context.Background()`. Lookups must keep auth, tenant, headers and processing mode.
- **S3. Stop `WithDryRun` from enabling mocks.** Mocks are a test concern. Make them explicit (`WithMockResponses(true)`) and update `DEFERRED_EXECUTION.md`, which also has a stale table. If this is too disruptive, add `WithDryRunOnly`, but decide before the tag.
- **S5. Uniform references.**
  - **(a) Ref-typed id parameters:** every `Get/Update/Delete` id parameter takes a ref type with one resolution contract. Plain ids pass through at no cost, so numeric-only resources (events, alarms) lose nothing.
  - **(b) One `ResolveID` signature:** tenant or parent scope moves into the ref or into the options.
  - **(c) Shared parser:** `source.ParseRef(s)` implements the CLI convention: `name:`/`ext:`/`query:`/`id:` prefixes are kept as given, all-digit values become ids, anything else becomes `name:`.
  - **(d) Raw bodies:** `CreateRaw`/`UpdateRaw` resolve the service's declared reference paths (`source.id`; `deviceId`/`agentId` for operations; `groupId` for bulk operations) when the value carries a ref scheme. Numeric values are left alone. A table per service declares the paths, which also documents them.
  - **(e) `ListAll` resolves once:** resolve option refs in `ListAll`, not on every page.
- **S6. Fix the method shapes where it is cheap.**
  - Make `Get`/`Delete` with and without options consistent.
  - Rename `binaries.Get` to `Download`, and add a metadata `Get`.
  - Fold `events.ListV2` into `List`.
  - Align the `...All` names.
  - The CLI codegen inversion proposal makes `ListOptions` the input for scaffolding, so the shape of the options structs is effectively public CLI surface. Run apidiff (§5.9 of the stabilization review) right after these changes.
- **S8. `op.Result` and `Iterator` as document sources.**
  - **Interface:** `interface{ Docs(ctx) output.Seq }`. A deferred or dry-run result yields nothing, a 204 yields nothing, `T` with `Unwrapper` yields its document, a plain value is JSON-encoded, and a binary result streams.
  - **Raw pages:** the iterator offers `Docs` and `PageDocs` (for `--raw`).
  - **Effect:** the CLI Runner then accepts `any` and adapts it in one place.

**Can land after the tag (additive):**

- **S4. Request inspection formatters.**
  - **API:** `api/inspect.Format(req, inspect.Curl|Markdown|JSON|Dump, inspect.Redact(...))`.
  - **Redaction:** covers Basic and Bearer tokens, cookies and `X-XSRF-TOKEN` (fixes C1).
  - **Parity:** port the v1 `PrintRequestDetails` output exactly, so the CLI golden tests hold.
  - **Effect:** the CLI no longer needs the v1 request handler.
- **S7. Pagination.**
  - Honor `CurrentPage` as the start page in `ListAll`.
  - Each service declares its default strategy, so commands never set `Strategy` unless the user passes `--paginationStrategy`.
  - Set `FullWalk` from `MaxItems == 0` instead of a separate flag.
- **S9. Public per-call hook.** `api.WithRequestHook(ctx, func(*http.Request) error)` handles the long tail (`--queryTemplate`, one-off headers) without new context options.
- **S10. Fill the remaining option gaps.** `withGroups` in `managedobjects.GetOptions` and `text` in binaries `ListOptions`. Optionally, a `CacheTransport` option for the response cache with the gzip and oauth exclusions now in `Runner.Client()`.

## 5. Required CLI changes (go-c8y-cli)

- **C1. A single `Call` type.** `Call func(ctx) any`. `Runner.Run` adapts with S8, installs the approver (S1, prompt plus yes/no to all from `submission.go`), and opens one approval scope per item.
  - **Delete:** `Submit*`, `ListCall` and the `From*` adapters in commands.
  - **Fixes:** this fixes the `tenantoptions/updatebulk` bypass by construction.
- **C2. Central pagination.** `r.Pagination()` builds `PaginationOptions` from the common options in one place, including `--raw`, `MaxItems`, `--includeAll` and `--paginationStrategy`. `ListAll` vs pages is decided by the adapter.
  - **Delete:** the 57 literals and the 17 copies of `applyQueryTemplate`. Move the query template into `r.Query` as an option.
- **C3. Flag kinds.**
  - **API:** `flags.Ref(cmd, "device", flags.DeviceRef, flags.Multi)` declares the flag type, completion, pipeline aliases and help text in one call. The same kind exists for group, agent, application, user, software version and firmware version.
  - **Behavior:** with `Multi`, a ref flag is always `StringSlice`, and fan-out works the same on every command (one item per value when the flag drives iteration; cycled override when stdin drives). `in.Ref(flag)` returns a normalized `source.Ref` using `ParseRef` from S5c.
  - **Delete:** `NameOrID`, `FirstID`, `groupRef`/`versionRef`/`patchRef` and the like.
- **C4. Never resolve in Build.**
  - **Lazy only:** pass refs to the SDK and let it resolve (S5).
  - **Delete:** `in.ResolveContext`, `ResolveSourceID`, `ResolveBodyRef` and the resolver closures, once S5d lands.
  - **Exception:** when one resolved id feeds several calls, resolve once with the public resolver.
- **C5. `r.Body` defaults the common getters.** It always includes the data, template and template-variable getters. A command lists only its own fields.
- **C6. `Factory.ClientV2()`.**
  - **Construction:** build the v2 client once per process, from the session. This moves cache, logging, the activity-log, output-template-capture and success-message middleware, and auth-type mirroring out of `Runner.Client()`.
  - **Completions:** switch the completion callbacks to the v2 client, which removes v1 imports from 202 files.
  - **Order:** do this after the holdouts are on v2, so credentials have a single owner.
- **C7. Migrate the v1 holdouts (section 1).**
  - Realtime subscribe on the SDK's notification2/realtime support.
  - `wait`/`assert` as Runner commands that poll typed `Get`.
  - `api` on `client.SendRequest` with `PrepareRequest`.
  - Then remove `github.com/reubenmiller/go-c8y` v1 from `go.mod`.
- **C8. Guardrail tests over the command tree** (`internal/clisurface`). Write them first and allow the current violations, so drift stops growing during the refactor:
  - **Mutating commands:** every command whose semantic method is POST/PUT/DELETE has a mode guard (`*ModeEnabled`).
  - **List commands:** every list command has `collectionProperty` and the common pagination flags.
  - **Flag types:** a flag name used across commands (`device`, `id`, `group`, `agent`) has the same pflag type everywhere.
  - **Banned identifiers in `pkg/cmd`:** a source-level check fails on `go-c8y/pkg/c8y"`, `pagination.PaginationOptions{`, `ResolveContext(`, `Submit`, `IsDryRun(`.
  - **Dry-run golden test:** every command runs once under `--dry` against `c8y-fakeserver` and must produce exactly one rendered request (or a declared count), never a real send.
- **C9. Then scaffold, not before.** The SDK-driven scaffold generator from `CLI_CODEGEN_INVERSION.md` should emit the shape from section 3. Generating today's shape would spread the boilerplate that C1 to C5 remove.

## 6. Suggested sequence

| Phase | Work | Why this order |
|---|---|---|
| 0 (now) | Fix `tenantoptions/updatebulk` (use `SubmitValue`). Add the C8 guardrails with an allow-list of current violations. Decide B1 (`withChildren`) from the remaining-issues doc. | Stops new drift and closes the one confirmation bypass. |
| 1 (before the v2 tag) | SDK S1, S2, S3, S5, S6, S8 | All are breaking or change behavior. After the tag they cost a v3 or deprecation shims. |
| 2 | CLI C1, C2, C5 in the Runner, then a mechanical sweep of `pkg/cmd` (gofmt/rewrite or codemod, one commit per group) | The Runner change is small; the sweep mostly deletes code. Expect each command file to shrink by 30 to 50%. |
| 3 | SDK S4, S7; CLI C3, C4 (ref flag kinds, lazy resolution only) | Needs S5. Removes the CLI's remaining Cumulocity knowledge. |
| 4 | CLI C7, C6; drop the v1 dependency | Holdouts are self-contained. Credentials get a single owner last. |
| 5 | Remove the guardrail allow-list entries; C9 scaffolding; flip `gen-powershell`/`gen-tests` to the tree | The surface is uniform, so projections and scaffolds stay uniform too. |

## 7. Open decisions

1. **Approval hook vs deferred prepare (S1).** I recommend the hook: it removes the double send, the multipart hang workaround and the four wrappers. The cost is a new public SDK API before the tag.
2. **Implicit body ref resolution (S5d).** Should `CreateRaw` resolve scheme-prefixed values at declared paths, or should the CLI switch to typed `Create(ctx, CreateOptions)` with `Fragments` for the templated rest? Typed is purer but loses the "body is whatever jsonnet produced" property the CLI relies on for `--template`. I recommend declared paths in the SDK.
3. **Mocks with dry run (S3).** Is breaking the implicit-mock default acceptable for existing v2 pre-release users?
4. **Multi-value `--device` everywhere (C3).** Should every create command fan out, or should `Multi` be opt-in per command? Either is consistent if the flag kind enforces it. Today it is neither.
5. **`withChildren=false` (B1).** I recommend re-baselining the goldens. The `*bool` refactor touches about 20 commands for no change on the wire.

## 8. Progress on the v3 branch

The `v3` branch of go-c8y-cli starts from `origin/v2`, merged with the v2 SDK migration from `exp-go-c8y-v2-output-filters` (including its staged work). The SDK changes are on the go-c8y v2 branch `feat/cli-v3-seams`.

### Done

| Item | Where |
|---|---|
| S1 request approver (`api.WithRequestApprover`, `WithApprovalScope`, `ErrRequestNotApproved`) | go-c8y `api/context_options.go`, `internal/contexthelpers/approval.go`, transport in `api/middleware.go` |
| S2 `ResolutionContext` keeps auth, tenant, headers and processing mode; public as `api.ResolutionContext` | go-c8y `internal/contexthelpers/deferred.go` |
| S3 (non-breaking variant) an explicit `WithMockResponses(false)` survives a later `WithDryRun`; the call order no longer matters | go-c8y `internal/contexthelpers/dryrun.go` |
| S5c `source.ParseRef` / `source.IsID` | go-c8y `api/source/parse.go` |
| S5d `DeviceResolver.ResolveBodyRefs`, used by `CreateRaw` of events, alarms, measurements and operations | go-c8y `api/inventory/managedobjects/body_refs.go` |
| S5e `ListAll` resolves device references once per walk | go-c8y events, alarms, measurements, operations |
| S7 (part) `ListAll` honors `CurrentPage` (offset strategy) | go-c8y `api/pagination/strategy.go` |
| S8 `op.Result.Documents`, `Iterator.Documents/PageDocuments` (`op.DocumentSource`) | go-c8y `op/documents.go`, `api/pagination/documents.go` |
| C1 single `Call` shape: `func(ctx) any`, rendered by `Runner.adapt`; confirmation through the SDK approver, one approval scope per item | `pkg/c8ystream/{runner,adapt,submission}.go` |
| C2 `Runner.Pagination()`; all 57 pagination literals and every `ListCall` removed | `pkg/c8ystream/pagination.go`, `pkg/cmd/**` |
| C4 (part) body device references resolved by the SDK for events, alarms, measurements, operations and the operation-creating commands (configuration send, software/firmware install and uninstall) | `pkg/cmd/**` |
| C8 guardrail tests with a shrinking allowlist | `internal/guardrails` |
| `tenantoptions updateBulk` confirmation bypass | fixed by C1 (no command can skip confirmation) |
| Uploads are confirmed against the real request (no `SubmitUpload` stub, no hang) | fixed by S1 and C1 |
| `NameOrID` and `ResolveContext` delegate to the SDK (`ParseRef`, `ResolutionContext`); scheme prefixes such as `ext:` are now kept | `pkg/c8ystream/resolver.go` |
| Commands that hand a file path to the SDK check `C8Y_DENIED_PATHS` (`Resolver.File`) | `pkg/c8ystream/resolver.go` |
| Wait and assert commands (14) on the Runner: one polling helper, `Runner.Assert` for pass-through, `--strict` and exit code 112, one shared assert factory (`pkg/c8ywaiter`, `pkg/desiredstate` and the duplicate devices factory removed) | `pkg/c8ywait`, `pkg/c8ystream/assert.go`, `pkg/cmd/inventory/assert/factory` |
| Certificate authority commands (get, create, update, delete) on the Runner; create settings, update and dry runs with the real fingerprint are in the SDK | go-c8y `trustedcertificates/certificateauthority`, `pkg/cmd/devicemanagement/certificate_authority` |
| The SDK's HTTP layer logs through `ClientOptions.Logger`; its warnings (credentials over plain HTTP) only show with `--verbose` | go-c8y `api/resty_logger.go`, `pkg/c8ystream/runner.go` |

### Pull requests (in review)

| PR | What |
|---|---|
| reubenmiller/go-c8y-cli#776 | body references and binary downloads handled by the SDK/Runner, `--id` takes several values everywhere |
| reubenmiller/go-c8y-cli#777 | one Go package per command group (359 packages merged into 87), no surface change |
| reubenmiller/go-c8y-cli#782 | realtime subscribe commands on the Runner, subscribe tests run offline |
| reubenmiller/go-c8y-cli#787 | wait and assert commands on the Runner and go-c8y v2 |
| reubenmiller/go-c8y#207 | certificate authority `Update`, create settings, complete dry runs; `*core.Error` fixes |
| reubenmiller/go-c8y-cli#788 | certificate authority commands on the Runner (stacked on #787) |
| reubenmiller/go-c8y-cli#789 | `inventory parents get` on the Runner (stacked on #788) |
| reubenmiller/go-c8y-cli#790 | `microservices serviceusers get` on the Runner (stacked on #789) |
| reubenmiller/go-c8y#211 | `UIApplicationPlugins` manages `config.remotes` (the model the UI uses) |
| go-c8y-cli `feat/v3-ui-plugins` | `ui applications plugins` list, install, delete, replace and update on the Runner (stacked on #790) |

Merged: go-c8y #199 (SDK seams), #200 (`op.Result` with 15 instead of 30 methods, the method/function rule and a compile-cost budget test), #202 (committed binaries removed), #204 (`ClientOptions.Logger`), #205 (fake server removes `null` fragments), #207 (certificate authority), #208 (fake server `withParents`), #209 (`ListServiceUsers`, collection results yield their items); go-c8y-cli #775 (CI for `v3`).

### Build time

Importing the go-c8y v2 `api` package makes every importing package compile the methods of every `op.Result[T]` and iterator instantiation (Go compiles generic instantiations per package). With about 300 command packages a cold build was 7 times slower than with go-c8y v1. Cold `go build -trimpath ./cmd/c8y`, one target:

| | CPU-seconds |
|---|---:|
| `origin/v2` (go-c8y v1) | 128 |
| v3 before | 914 |
| go-c8y#200 (as first proposed, 12 methods) | 727 |
| go-c8y-cli#777 | 313 |
| both | 237 |

The guardrail tests make the per-package cost visible again if new command packages are added; further gains would come from fewer pagination iterator and realtime stream methods (about 400 functions per importer).

### Still open

The allowlist in `internal/guardrails/allowlist.json` is the worklist. Each entry disappears when it is fixed. After the PRs above:

- `go-c8y-v1` (mostly completion callbacks) and `v1-worker`: bulk registration, enroll, hosted applications, microservice create, `api`, `assert text`, remote access, sessions (C6, C7).
- `measurements createBulk` groups several piped items into one request. The Runner makes one request per item, so it needs a batching option in the Runner first.
- `alarms assert count --fragmentType` is ignored: `alarms.ListOptions` has no fragment type (the alarm API does not document one).
- The fake server reports `totalPages: 1` for an empty collection (Cumulocity reports 0). The count asserts now use `totalElements`, so they are not affected.
- `flag-type-*`: `--id` differs in 3 commands (tenants create is intentional), `--group` in 3 (bulk registration), `--device` in 1 (remote access configurations update).
- `dry-run-check` (2): bulk registration and enroll, both still v1.
- `mode-guard` (9): local commands (alias, cache, sessions, settings, extensions); probably legitimate exceptions to mark as such.
- C5 (`r.Body` defaults), S4 (request formatters in the SDK), S6 (method signatures), S9, S10.
- The `--template` piped-input fix for software/firmware version create from `origin/v2` (#378) is not yet ported to the Runner commands.
- Several piped devices on `subscribe` are subscribed to one after the other; subscribing to all of them on one stream would match the intent better.
