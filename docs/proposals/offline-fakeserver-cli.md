# Offline fake Cumulocity server for CLI tests

Status: proposal. Goal: give go-c8y-cli (and ongoing go-c8y v2 SDK changes) a fast,
hermetic feedback loop that never touches a real Cumulocity tenant — by reusing the
**stateful fake server that already exists in go-c8y v2**.

Related: [v2-integration.md](./v2-integration.md) (the SDK migration this supports),
`go-c8y-v2/docs/OFFLINE_TESTING.md` (the fake server's own docs).

## Motivation

The CLI test suite splits in two today:

- **`tests/auto/`** runs in dry-run (`C8Y_SETTINGS_DEFAULTS_DRY=true`). It asserts the
  request the CLI *would* send (method/path/body) and never contacts a server. Good
  coverage of request construction, **zero coverage of response handling**.
- **`tests/manual/`** runs real shell flows — create → wait-for-fragment → delete,
  pagination, pipe chains that feed real IDs forward — and **requires a live tenant**
  (`C8Y_HOST`/`C8Y_USER`/`C8Y_PASSWORD`). Slow, order-sensitive, pollutes a real tenant,
  and is the source of the session-gated golden failures (`template`, `common/select`,
  `common/pipe`).

During a large refactor we want to validate *actual round-trip behaviour* (the response
parsing, paging, select/filter pipeline, name→id resolution) offline and instantly. A
simulated server is the missing piece.

## The asset already exists

`go-c8y-v2/internal/pkg/fakeserver/` (~5,700 LOC) is a stateful, in-memory fake
Cumulocity platform backed by `httptest.Server`:

- CRUD + filtering + pagination with the real collection envelope
  (`statistics`/`self`/`next`) for inventory, identity, measurements, events, alarms,
  operations/bulk, applications, binaries, users/groups/roles, tenants/options, audit,
  retention, login options, device requests, certs; WebSocket handlers for realtime &
  notification2.
- Seed data every tenant has (`t12345`, `admin`, login options, system version
  `1016.0.0`, cockpit/devicemanagement apps, sample devices/alarms).
- Three modes via `TEST_MODE`: `offline` (fake), `live` (real), `record` (proxy a real
  server → golden files) + a golden-comparison validator. **378 pass / 20 skip in ~13s.**

This is the simulated server. The work is not to build it — it is to make it reachable
from the CLI and wire it into the manual harness.

## The one structural constraint

```go
// go-c8y-v2/internal/pkg/fakeserver/server.go
func New(t *testing.T) *FakeServer { ... t.Cleanup(fs.Server.Close) ... }
```

Two facts follow:

1. **`internal/` is not importable across modules.** The CLI is a separate module; even
   with `replace github.com/reubenmiller/go-c8y/v2 => ../go-c8y-v2` (already in the CLI
   `go.mod`), `internal/pkg/fakeserver` stays off-limits. `replace` redirects the module
   source; it does **not** lift the `internal/` visibility rule.
2. **`New(t)` is coupled to `testing`.** It can only run inside a Go test.

The key realisation: the CLI's manual tests run the `c8y` **binary as a subprocess**, so
they need a *real listening server over HTTP*, not a Go import. `httptest.Server` already
listens on a real localhost port (`fs.Server.URL` → `http://127.0.0.1:<port>`). A `main`
package **inside** go-c8y-v2 may import `internal/`. So the out-of-process path sidesteps
constraint #1 entirely — no package move, no module publish.

## Design — two reuse modes

```
                    go-c8y-v2 fakeserver  (httptest.Server = real localhost port)
                              ▲                                  ▲
        in-process import     │                                  │   real HTTP over 127.0.0.1:PORT
   (v2 SDK tests; future  ────┘                                  └──── c8y binary as a subprocess
    Go-level CLI tests over                                            (commander manual/golden suite)
    the c8ystream v2 path)                                       ── needs only a standalone binary ──
   ── needs promotion to pkg/ ──                                    (no internal/ move, no publish)
```

**Recommended primary path: out-of-process.** A tiny `cmd/c8y-fakeserver` in go-c8y-v2
boots the fake, prints its URL, and blocks on SIGINT. The CLI harness points a throwaway
session at it. This is the high-leverage change and requires no public API surface.

**Secondary path: in-process.** If/when we add Go-level CLI tests that drive the v2 client
directly (the `c8ystream` bridge), promote `fakeserver` to `pkg/` and import `srv.URL()`.
Defer until needed — it commits us to maintaining the fake as public API.

## Changes required

### go-c8y-v2 (enabler)

1. **Decouple the constructor from `testing`.** Split the core out; keep the test
   ergonomics as a thin wrapper.

   ```go
   // pure constructor — usable from main() and tests
   func NewServer() *FakeServer { /* build stores, mux, httptest.NewServer, seed */ }
   func (fs *FakeServer) Close() { fs.Server.Close() }

   // test wrapper preserves the existing call site ergonomics
   func New(tb testing.TB) *FakeServer {
       fs := NewServer()
       tb.Cleanup(fs.Close)
       return fs
   }
   ```
   `testing.TB` (not `*testing.T`) so benchmarks/fuzz can use it. The `testcore` factory
   and ~400 test call sites are unaffected (`New(t)` keeps working).

2. **Add `cmd/c8y-fakeserver/main.go`.** Flags: `--addr`/`--port` (fixed port for the
   harness, else ephemeral), `--seed` (default on), optionally `--tls`. Start `NewServer()`,
   print the resolved base URL to stdout (one line, machine-readable), block until SIGINT,
   then `Close()`. `main` imports `internal/pkg/fakeserver` legally.

3. *(Deferred — only for the in-process path)* move `internal/pkg/fakeserver` →
   `pkg/fakeserver`. Decision gate in Open Questions.

### go-c8y-cli (consumer)

No production code changes. The session is env-only (host + credentials), exactly how CI
drives the manual suite today; the fake only checks that an `Authorization` header is
*present* (value ignored), so any non-empty credentials authenticate.

- Add `tests/run-manual-offline.sh` (wrapper around `run-manual.sh`):
  1. `go run ../go-c8y-v2/cmd/c8y-fakeserver --port 0` (or build it once), capture the URL.
  2. Export the session env the CLI reads — `C8Y_HOST=<url>`, `C8Y_TENANT=t12345`,
     `C8Y_USER=admin`, `C8Y_PASSWORD=admin` — pointed at the fake.
  3. **Run `tests/scripts/setup.sh` against the fake.** It already provisions fixtures
     *with `c8y` itself* (`create_user`, `create_agent`, `create_device_with_service` …),
     so it populates the fake with the exact fixtures the manual tests expect — no need to
     re-encode them as Go seed data.
  4. `commander test --config ./config.manual.yaml --dir manual/$folder` (unchanged;
     `config.manual.yaml` already has `inherit-env: true`).
  5. Trap-kill the fake server on exit.
- Add `task test-cli-offline` mirroring `test-cli-manual`.

## Gaps & risks (honest)

| Risk | Detail | Mitigation |
|---|---|---|
| **CQL/`q` filtering depth** | `filtering.go` models a subset (`has()`, `field eq 'value'`, `and`). The CLI leans hard on inventory queries. | Fetch-by-id and request-assertion tests unaffected; extend `filtering.go` as list-filter assertions surface gaps. Biggest single time sink. |
| **CLI request drift** | Fake was built for the *SDK's* request shapes; the CLI emits shapes the SDK tests don't (`withTotalPages`, page-size edges, template-built bodies). | Use `record` mode against a live tenant to capture CLI traffic, diff against the fake, fill `handler_*.go`. Additive. |
| **Shared-server state** | Whole manual run shares one fake (like today's shared live tenant), vs the SDK's fresh-per-test server. | Acceptable (matches current manual model). Offer per-group servers later if isolation bites. |
| **Auth realism** | Header-presence only; no real OAuth/token/2FA, no per-user permission enforcement. | Fine for functional CLI tests. Auth-specific flows stay live or skipped. |
| **Out of scope (the 20 skips)** | Realtime over real WS, external-URL binary downloads, microservice Docker upload, CA ops. | Remain live-only; subscribe/notify still works via the WS handlers. |
| **Session bootstrap** | CLI login probe (currentTenant/loginOptions/version) must be satisfied by seed data. | Seeds cover these; verify with a 5-min smoke test before trusting the suite. |

## Phased plan

- **Phase 0 — enabler (go-c8y-v2, ~½ day).** Decouple `New`, add `Close()`, add
  `cmd/c8y-fakeserver`. Existing SDK tests stay green.
- **Phase 1 — thin slice (CLI, ~½ day).** `run-manual-offline.sh` + `task test-cli-offline`;
  prove **one** manual group (e.g. `inventory` or `common/pipe`) green offline end-to-end,
  including `setup.sh` provisioning against the fake.
- **Phase 2 — broaden + guardrail (open-ended).** Run the full manual suite offline; triage
  failures into (a) fake handler/filter gaps → extend, (b) real CLI regressions → fix. Use
  `record` mode to keep the fake honest against a live server.
- **Phase 3 — optional in-process.** Promote `fakeserver` to `pkg/`; add Go-level CLI tests
  over the `c8ystream` v2 path with an in-process server.
- **Phase 4 — optional.** Convert select dry-run `auto` tests into real-response round-trip
  tests where response behaviour (not just the request) is what matters.

## Open questions

1. **Long-term home of `fakeserver`:** stays `internal/` + binary only (test infra, no
   public API commitment), or promoted to `pkg/` as a supported, versioned artifact? Phase 3
   forces this; Phases 0–2 do not.
2. **Distribute the binary?** Test-only, or shipped so CLI/automation users can dry-test
   their own scripts against a fake tenant?
3. **Server lifecycle for the CLI suite:** one shared fake per run (proposed) vs per-group
   reset, weighed against suite runtime and cross-test coupling.
4. **Fixture source of truth:** `setup.sh`-against-fake (proposed, reuses existing
   provisioning) vs richer built-in seed data in the fake. The former keeps one definition.

## How the suites run today (v3)

| What | Command | Where |
|---|---|---|
| Go tests (`go test ./...`; `cmd/tests` against the fake server) | `task test-go-offline` | CI `test-offline` |
| Dry-run tests (`tests/auto`) | `task test-cli-offline` | CI `test-offline` |
| Realtime tests | `task test-cli-offline-realtime` | CI `test-offline` |
| Manual tests the fake server supports | `task test-cli-offline-manual` | CI `test-offline` |
| Live tenant (auto, manual, `cmd/tests`) | `gh workflow run test-live.yml --ref v3` | `.github/workflows/test-live.yml`, on demand |

`cmd/tests` skips when `C8Y_HOST` is not set (`command.RequireTenant`), so a plain
`go test ./...` without a tenant passes. `tests/run-offline.sh` with
`C8Y_OFFLINE_SUITE=go` starts the fake server, creates the test data with
`tests/scripts/setup.sh` and runs `go test` with the fake server configured.
