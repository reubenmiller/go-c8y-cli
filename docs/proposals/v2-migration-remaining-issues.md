# go-c8y v2 migration — remaining manual-test issues

> **Update 2026-06-20 — section A fixed.** A1–A6 are implemented (see each
> section's *Resolution* note). Verified offline via `util show` / `--dry` where
> possible; the parts that need a tenant (A1's full 9-test set, A6's exit-code,
> the `--data`-only round-trips) are noted per item. Section **B1** is still a
> decision (re-baseline vs `*bool`) and section **C** remains verification-only.

Status snapshot after the c8ystream regression-fix work. The original
`test-cli-manual.log` reported **148 failures** out of 866, all regressions from
the v2 (`pkg/c8ystream`) migration. The large clusters (pagination defaults,
inventory-query grouping, templates, caching, `--select`/shaping, pipe-mapping,
multipart, custom headers/query params, `--noAccept`, plus a long tail of misc
fixes) are resolved. This document captures what is **left**, the root cause for
each, and advice for finishing it.

## How to read this

- Paths without a repo prefix are in **go-c8y-cli**; paths beginning `../go-c8y-v2`
  are in the **go-c8y v2 SDK** (wired via the `replace` directive in `go.mod`).
- **Verification** tags:
  - 🟢 *offline* — reproducible with `--dry`/templates/`util show` and a dummy
    session (`C8Y_HOST=https://example.com …`). Fixable and verifiable here.
  - 🔴 *needs-tenant* — the test makes a real request (name resolution, repository
    lookup, round-trip create/update/delete, `--dry=false`, `--verbose`, a 2xx
    body). The underlying logic can be reasoned/fixed offline but only **verified**
    against a live Cumulocity tenant.

---

## A. Offline-fixable remaining

### A1. `--outputTemplate` only binds `output`/`index` (8 + 1 tests) 🔴

**Tests:** `outputTemplate.yaml` — *maps input.value to the object*, *provides
access to response and flags*, *sets input to null if no iteration values are
used*, *supports outputting toml output*, *supports reading the output template
from a file*, *supports table output*, *supports templates producing string
output*, *preserves leading and trailing whitespace*; `flag_withError.yaml` —
*should allow output templates to use the error when shaping the output*.

**Root cause:** the v2 output-template stage is built in
`CommonStages` ([pkg/c8ystream/render.go](../../pkg/c8ystream/render.go), the
`cfg.GetOutputTemplate()` block) via `template.Jsonnet(tpl)`. That wrapper
(`../go-c8y-v2/pkg/c8y/output/template/jsonnet.go`, ~line 60) binds **only**
`output_json` and `index`. The v1 path (`pkg/mapbuilder/outputtemplate.go`,
`outputTemplateHeader()`) also binds `input`, `flags`, `request`, `response`,
`rand`, `time` and the `_` helper library. Any template referencing
`input.value`, `request.*`, `response.*`, `flags.*` therefore resolves to nothing
and the document lacks the expected field.

A second, related defect (RC4): a template whose result is a **top-level JSON
string** (e.g. `std.join(',', …)`, `std.manifestTomlEx(…)`) is printed quoted /
escaped instead of unwrapped. v1 unwraps it in `pkg/request/request.go` (~921‑940:
when `gjson` `Type == String`, emit `parsedOutput.Str`).

**Advice:**
1. Route the v2 output template through `mapbuilder.OutputTemplate` (or bind the
   same ext-vars on the jsonnet VM). Because the stage in `CommonStages` is
   context-free, the runner must thread the per-item `input` (`in.input`) and,
   where available, `request`/`response` metadata into the evaluation — mirror
   `console.ExecuteTemplate` (`pkg/request/request.go` ~709‑735) which builds the
   `requestData`/`responseData` maps and `commonOptions.CommandFlags`.
2. **Minimal interim** that fixes the `input`-only and `flags`-only tests without
   request/response plumbing: bind `input` from the resolver and `flags` from the
   command's changed flags.
3. For RC4, after the template stage, if `gjson.ParseBytes(out).Type ==
   gjson.String` emit the decoded `.Str`; numbers/bools/null pass through as-is.

**Verification:** all 8 use `devices get agent01` / `alarms count` / `devices
list` — needs a tenant.

**Resolution (2026-06-20):** the v2 output-template stage is now built by the
runner (`pkg/c8ystream/outputtemplate.go`, `Runner.outputTemplateStage`) routed
through `mapbuilder.OutputTemplate`, so it binds the full v1 environment
(`_`/`vars`/`var`/`rand`/`time`/`input`/`flags`, plus `request`/`response`). The
per-item `input` is threaded via `Runner.tmplInput` (set in the producer; correct
for the serial pipeline). `flags` comes from `Runner.commandFlags()` (changed
flags). RC4 is handled by `unwrapTemplateString` (top-level `gjson.String` →
`.Str`). `CommonStages` now takes the stage as a parameter so ordering stays
filter→template→select. Verified offline through `util show --outputTemplate`:
`input.value`, `flags`, the `_` library, top-level string unwrap and
leading/trailing whitespace preservation all work. **Still open:**
`request`/`response` are bound empty (the two tests that read them —
*provides access to response and flags*, `flag_withError` — need per-item
`op.Result` plumbing and a tenant). The toml test's leading `\n\n` is a
go-jsonnet `std.manifestTomlEx` stdlib artifact, not the binding.

---

### A2. CSV `--select` into arrays returns the whole array (2 tests) 🟢

**Tests:** `measurements_createBulk.yaml` — *can override the timestamp of the
piped input*, *uses time from piped input if an override is not defined*
(`--select "body.measurements.**.time" --output csv`).

**Root cause:** the select stage (`selectStage` in
`pkg/c8ystream/render.go`) emits nested JSON via
`Selection.JSON()`. The CSV encoder derives its columns from
`leafPaths(firstDoc)` (`../go-c8y-v2/pkg/c8y/output/encode/csv.go`, ~line 87),
which **recurses only into objects** — arrays are treated as leaves. So
`body.measurements` becomes a single column holding the entire array JSON
(`"[{""time"":…}]"`) instead of `body.measurements.0.time`.

**Advice (in order of preference):**
- **(c)** For the `--select` + CSV/TSV case, render from `Selection.CSV(sep)`
  directly (the `Selection` already knows the flat keys incl. array indices and
  stable empty columns) — this is exactly the v1 `pkg/jsonfilter` path. Requires
  bypassing `encode.CSV` when `--select` is active.
- **(a)** Pass the selection's keys to `encode.CSV` via `CSVOptions.Columns`
  (caveat: `**` makes columns vary per document).
- **(b)** Make `leafPaths` recurse arrays — **risky**: changes *all* CSV output
  (e.g. `c8y_SupportedOperations: [...]` would become indexed columns), likely
  breaking currently-passing CSV tests.

This also fixes the **stable-empty-column** gap in `select.yaml` (e.g. *includes
empty values for non-existent values in the last field* — a trailing unmatched
`--select` column should emit an empty CSV cell), which is currently
🔴 needs-tenant but shares this root cause.

**Resolution (2026-06-20):** implemented option (c). A selection-aware CSV
renderer (`pkg/c8ystream/csvselect.go`, `selectionCSVRenderer`) renders
`Selection.CSV(sep)` directly; `CommonStages` skips the generic `selectStage`
when `selectCSVActive(cfg)` (CSV/TSV + `--select`), and `--outputFile` is shaped
via a `shapingRenderer`. Also fixed `shape.Selection.CSV` to use
`marshalNoEscapeHTML` (it HTML-escaped `&`/`<`/`>` → e.g. a `pageSize=20&q=…`
cell became `…&…`). Both createBulk tests and the trailing-empty-column
test pass; verified offline.

---

### A3. `measurements create` lost multi-device fan-out (1 test) 🟢

**Test:** `measurements_create.yaml` — *creates multiple measurements via
pipeline but can override with multiple devices* (`--device 11111,22222`).

**Root cause:** `pkg/cmd/measurements/create/create.go:54` declares
`cmd.Flags().String("device", …)` — a single string. v1 used `StringSlice`. So
`11111,22222` is taken as one value and either resolves the bogus name
`"11111,22222"` or yields 0 outputs; the per-device fan-out is lost.

**Advice:** restore `StringSlice("device", …)` plus the slice/override device
handling (mirror events/alarms create, which fan out one item per device via the
override iterator + `c8yfetcher` device resolver). Other source-bearing creates
(`events create`, `alarms updateCollection`) have the same single-`String`
downgrade and need the same change *only if* multi-device is required there.

**Resolution (2026-06-20):** `device` is now `StringSlice` in
`measurements/create`. A new `Resolver.ApplyOverrideSlice(body, flag, path)`
(`pkg/c8ystream/resolver.go`) cycles the slice one value per piped item by
`index % len` (v1's override-iterator fan-out), overriding the `source.id` that
`WithOverrideValue` seeds for the required-property check. `WithStringValue` for
`device` was dropped (it errors on a slice flag). Both fan-out tests pass; the
remaining utf8 test is a real `devices create` (tenant).

---

### A4. `devices list --includeAll` dry-run hides the keyset request (1 test) 🟢

**Test:** `devices_list.yaml` — *optimizes the inventory query when using
retrieving all devices via includeAll flag*. Expected
`pageSize=2000&q=$filter=(_id gt '0') $orderby=_id asc`; got `pageSize=2000&q=`.

**Root cause:** the pagination dry-run fix renders the caller's **pristine base**
request (no keyset cursor, no forced totals) for every list — correct for a
normal single-page list, but `--includeAll` *wants* the id-keyset first request
shown. Normal and includeAll use the same id-keyset strategy; the only
distinguishing input is `--includeAll`.

**Advice:** thread a "full-walk" signal into `PaginateWith`
(`../go-c8y-v2/pkg/c8y/api/pagination/iterator.go`) — e.g. a `PaginationOptions`
field the CLI sets when `cfg.IncludeAll()`. When walking all pages, render the
dry-run from `strategy.First(base)` (keyset-decorated) minus the forced
`withTotalElements/withTotalPages`/`currentPage`; otherwise keep the pristine
base. Contained SDK change; ~1‑2 tests.

**Resolution (2026-06-20):** added `PaginationOptions.FullWalk bool` (`url:"-"`).
`PaginateWith` computes a `dryRunReq` = `strategy.First(dryRunBase)` with
`CurrentPage`/`WithTotalPages`/`WithTotalElements` reset to the base when
`FullWalk` is set, and renders the dry run from it (offset strategy is unchanged
since its `First` only sets currentPage, which is reset). `devices/list` sets
`FullWalk: includeAll`. The includeAll dry-run now shows
`pageSize=2000&q=$filter=(_id gt '0') $orderby=_id asc`; normal lists unchanged.
All 10 `devices/list` manual tests pass.

---

### A5. `binaries create` object JSON pretty-print under `--template` (1 test) 🟢

**Test:** `binaries_create.yaml` — *adds custom fields to the binary using a
template* (`--dryFormat markdown`).

**Root cause:** the markdown dry-run renders the multipart `object` JSON
**compact**, but this test (object built via `--template`) expects **pretty**
(3-space indent), while the *basename* test (`--data` only) expects compact. The
formatting difference correlates with `--template` usage — the template-built
body appears to retain pretty formatting that the compact path strips.

**Advice:** compare against v1's dry-run markdown multipart rendering
(`pkg/request/request.go`) to confirm whether v1 pretty-printed JSON parts or the
template body carried its own indentation, then make the v2 dry-run match. Low
priority (1 test). The multipart *content-type* logic for this test is already
correct.

**Resolution (2026-06-20):** added `core.UploadFileOptions.ObjectIndent string`;
`NewMultiPartFileFields` uses `json.MarshalIndent(obj, "", ObjectIndent)` when
set, else compact. `binaries/create` sets `ObjectIndent: "   "` when `--template`
was used, so the template object is pretty (3-space) and the `--data` object
stays compact. The template + basename + manual-name tests pass; the remaining
failure is the real create/update script (tenant).

---

### A6. `inventory find` double-encodes a pre-encoded query (1 real test) 🔴

**Test:** `queryEncoding.yaml` — *does not double encode query parameters*
(`inventory find --query "name eq 'salt %26 pepper'" --verbose`). The other 5
`queryEncoding` failures are live-server `--verbose` round-trips.

**Root cause:** `pkg/cmd/inventory/find/find.go:170` passes `q.Build()` (which may
contain a user-supplied pre-encoded `%26`) into `ListOptions.Query`; the v2 client
URL-encodes again → `%26` → `%2526`. v1 decoded once before the SDK encoded.

**Advice:** decode **once** before the SDK encodes — but `url.QueryUnescape`
turns `+` into a space, which would corrupt timestamp queries
(`…T00:00:00+02:00`). Use **`url.PathUnescape`** (decodes `%XX`, leaves `+`
alone) at `find.go:170`, or fix the SDK so it doesn't re-encode an
already-encoded query value. Verify against timestamp/`+` queries. The test
itself asserts `exit-code: 0`, so it needs a tenant to fully pass; the encoding
can be checked offline via `--dry`'s `pathEncoded`.

**Resolution (2026-06-20):** `inventory/find` decodes `q.Build()` once with
`url.PathUnescape` before the SDK re-encodes (invalid escapes fall back to the
raw query). Verified offline via `--dry`: pre-encoded `%26` → `salt+%26+pepper`
(no longer `%2526`), literal `&` still encodes, complex chars (`%$`) don't crash,
and a timestamp `+02:00` is preserved (`%2B`, not a space). All 23 auto inventory
tests pass; exit-code-0 still needs a tenant.

---

## B. Behavioural-change candidates (assess before "fixing")

These are differences where the **v2 output is arguably cleaner** but breaks a v1
golden. Functionally equivalent on the wire — decide whether to restore v1 or
re-baseline the golden.

### B1. `devicegroups list` omits `&withChildren=false` (3 tests)

**Tests:** `devicegroups_list.yaml` — *finds device groups by name*, *finds root
device groups*, *finds sub (nested) device groups*.

**Root cause:** `withChildren` lives on the shared
`managedobjects.GetOptions` (`../go-c8y-v2/.../inventory/managedobjects/api.go:171`)
as `bool url:"withChildren,omitempty"`. `omitempty` drops `false`. devicegroups
expects `withChildren=false` **always** sent; inventory/devices expect it
**omitted**. The shared `bool` cannot express both, and ~20 list commands set it
via `in.Bool("withChildren")`.

**Advice:** `withChildren=false` is functionally identical to omitting it (the
server default is `false`). Prefer **re-baselining** the 3 goldens. If you must
restore v1: change `WithChildren`/`WithChildrenCount` to `*bool` in `GetOptions`,
then have devicegroups use `in.BoolPtr(...)` (always send) and the other ~20
callers use `in.BoolPtrIfChanged(...)` (omit unless set) — high blast radius,
verify the whole inventory family.

---

## C. Live-server-only (logic fixed, needs a tenant to verify)

These were in the 148 and the underlying fix is in place, but they exercise a
real request so they can only be confirmed against a tenant.

| Area | Tests | Fix status / note |
|------|-------|-------------------|
| `--select` shaping | `select.yaml` (6), `select_context_shaping.yaml` (8), `select_flat_selection.yaml` (1) | Aliases/`**`/negation/empty-objects/HTML-escape fixed and verified offline with synthetic input; tests use `devices create`+`list` / `applications get`. |
| Caching | `cached_commands.yaml` (5), `cache_delete/renew/bodyPaths` | Cache RoundTripper now installed in `Runner.Client()`; needs a server to populate/read. |
| Firmware/software install | `firmware_versions_install`, `software_versions_install` | `version`/`url` default-to-`""` fixed; the no-`--url` path does a repository lookup that needs a server. |
| Round-trip create/update/delete (`ExitCode`) | `firmware_patches`, `firmware_versions_create`, `software`, `plugins`, `configuration` | Run real create→update→delete chains. |
| Errors | `errors_command` (timeout, multi-error), `errors_handling` (`--withError`) | `Timeout` now passed; producer errors continue; `core.Error.Response` tagged `json:"-"`. Remaining diffs depend on real server error bodies. |
| Output | `output_json_compact` (statistics totals) | Depends on a real collection response. |
| Sessions / login | `sessions_set`, `session_login` | Not migrated to c8ystream; need a live host for token/version. |
| Confirmation message | `alarms_updateCollection` (*✓ Updated … => 200 OK*) | **Not yet ported** — v1 emitted a success confirmation for no-content/204 responses (`pkg/request/request.go` ~768‑794 `showMessage`); the c8ystream submission/render path has no equivalent. Port the method→action map + ForceTTY gate. Needs a server for the `200 OK`. |
| Misc real requests | `pipe_applications` (file→json-lines), remaining `piping_collection` (2), `inventory_subscribe`, `api.yaml` (2, `--dry=false`), `flag_verbose` (1) | Make real calls; offline they 404 against `example.com`. |

### C1. `flag_dry_curl` Bearer-vs-Basic redaction (1 test) — environment artifact

`flag_dry_curl.yaml` — *displays the curl command for creating a binary*. In the
original live-cluster capture the curl showed a raw Bearer JWT instead of the
redacted `Basic {base64 …}` placeholder, because `HideSensitiveInformation`
(`pkg/config/cliConfiguration.go`) redacts `client.Token` via `strings.ReplaceAll`
and the legacy `*c8y.Client` used for redaction had `Token == ""` (the token was
acquired by the v2 client's token source). Offline (Basic auth) it redacts
correctly, so this is **not** a reproducible offline regression.

**Advice (separate hardening ticket):** propagate the resolved token onto the
legacy client used for redaction, or add a generic `Bearer [A-Za-z0-9._-]+`
redaction regex alongside the existing Basic-auth one.

---

## Suggested order of attack

1. **A2** (CSV array select) and **A3** (measurements multi-device) — small,
   offline-verifiable, self-contained.
2. **A1** (output-template env binding) — highest test count (9), but needs a
   tenant to verify; do the minimal `input`/`flags` binding first.
3. **C** sweep against a tenant to confirm the already-applied fixes, and port
   the **`showMessage` confirmation** (`alarms updateCollection` and the family of
   no-content commands).
4. **A4/A5/A6** — niche, 1 test each.
5. **B1** — decide re-baseline vs `*bool` refactor.
