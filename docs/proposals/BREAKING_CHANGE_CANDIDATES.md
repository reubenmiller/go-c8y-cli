# Breaking-change candidates (go-c8y-cli v3 and go-c8y v2)

This page collects improvements that would change existing behavior. Each one needs a decision before it ships, because it can break scripts, extensions or SDK callers.

Until a topic is decided, the current (v2 CLI) behavior is kept. Opt-in additions, such as a new flag or a new command that does not change any default, can ship without a decision.

Sources: go-c8y-cli PRs #777 and #797 to #811, go-c8y PRs #220 to #239, the go-c8y v2 CHANGELOG, the SDK consistency audit and the `tests/manual` changes between master and v3.

## How to read an entry

- **Today**: what users get now (v2 CLI, or the current v3 behavior where noted).
- **Proposed**: the better behavior.
- **Breaks**: who notices.
- **Ways to ship**: options that limit the damage, usually one or more of:
  1. opt-in first (flag, setting or environment variable), default unchanged;
  2. deprecation warning when the old behavior is relied on;
  3. flip the default in a major release, with the old behavior still available as an opt-out.

## Part 1: proposed improvements (decision needed)

### 1. Idempotent delete (missing id is not an error)

- **Today**: deleting an id that does not exist exits 4 (one item) or 104 (several items). Scripts silence it with `--silentStatusCodes 404` and `--silentExit`. The go-c8y v2 SDK treats a 404 on delete, unassign and unsubscribe as success (`IgnoreNotFound`); the CLI turns it back into an error with `api.WithNotFoundAsError` so v3 keeps the v2 exit codes.
- **Proposed**: "make sure it is gone" semantics. A missing id is success, as in the SDK, Terraform and most declarative tools. Cleanup scripts need no extra flags.
- **Breaks**: scripts that check exit 4/104 after a delete, or use the 404 as a signal that the object was already removed.
- **Ways to ship**: an opt-in setting (for example `settings.defaults.idempotentDelete` or a `--ignoreNotFound` flag), then a later default flip with `--failOnNotFound` as the opt-out. `--silentStatusCodes 404` keeps working either way.
- **Source**: CLI#806, go-c8y#234, `tests/manual/common/errors/test_delete.sh`.

### 2. Idempotent create and subscribe (409 conflict is not an error)

- **Today**: creating something that already exists (external id, child assignment, application subscription, notification2 subscription) fails with the 409 status. The SDK already treats these as success (`IgnoreConflict`, status Duplicate). Whether v3 currently keeps the v2 exit codes for them is being checked; if not, they will be restored like the 404 case.
- **Proposed**: a 409 on these "ensure" operations is success, with the existing object in the output.
- **Breaks**: scripts that use the 409 to detect "already exists".
- **Ways to ship**: same options as topic 1, ideally one shared setting for both (for example `--idempotent`).
- **Source**: go-c8y SDK `IgnoreConflict` call sites (identity create, child assign, applications and microservices subscribe, notification2 subscription create).

### 3. Create-or-get and upsert commands

- **Today**: the CLI has only create, get and update. Users combine `get` plus `create` in scripts, which races and needs error handling.
- **Proposed**: expose the SDK's idempotent operations as commands or flags. The SDK already has:
  - managed objects and devices: `GetOrCreateByName`, `GetOrCreateByFragment`, `GetOrCreateWith`, `GetOrCreateByExternalID` (creates the object and its external id in one request);
  - software: `GetOrCreateByName`, `GetOrCreateWith`; firmware: `GetOrCreate`/`UpsertByName`, versions `GetOrCreateVersion`/`UpsertByVersion`, patches `GetOrCreatePatch`/`UpsertByVersionAndDependency`; configuration `GetOrCreate`/`UpsertByName`;
  - alarms `Upsert`, event binaries `Upsert`, DTM assets and definitions `GetOrCreate`/`Upsert`.
- **Breaks**: nothing if added as new commands (for example `devices getOrCreate`, `inventory upsert`, `alarms upsert`) or new opt-in flags (`--ifNotExists`). It becomes breaking only if existing `create` commands change their default.
- **Decision**: naming (new verbs versus flags on create), how to report "created" versus "found" (output field, exit code, or both), and which commands first.

### 4. Websocket commands honor `--dry`

- **Today**: realtime subscribe commands and remote access (`remoteaccess connect ...`, `server`) run their lookups and then open the real connection under `--dry`.
- **Proposed**: under `--dry`, print the upgrade request and the channels, and do not connect. The SDK supports this since go-c8y#230.
- **Breaks**: anyone using `--dry` with these commands expecting a live session (unlikely but possible).

### 5. One documented exit-code table

- **Today**: 100 (unexpected error) and 101 (user error) are used inconsistently between built-in, extension and runtime commands. Each v3 PR documents its own corner (see part 2).
- **Proposed**: a single mapping (HTTP status codes, user errors, unresolved names, timeouts, assertion failures), documented and tested, before v3 is tagged.
- **Breaks**: scripts that match the exit codes that change as a result.

### 6. Name resolution: wildcards and ambiguous matches

- **Today (v3)**: user lookups in extension commands no longer expand wildcards (`--user 'adm*'`); the SDK's new `Users.ResolveID` supports wildcards but returns an "ambiguous" error when more than one user matches, where v1 took the first match. Role lookups keep wildcard and first-match.
- **Proposed**: one rule for all name lookups: wildcards allowed; more than one match is an error that lists the candidates (safer for update and delete), with an opt-in `--first` for the old behavior.
- **Breaks**: scripts that rely on the first match.

### 7. Login TLS verification

- **Today (kept on purpose)**: BASIC, OAUTH2_INTERNAL and OAUTH2 login and the login-options lookup accept self-signed certificates whatever `--allowInsecure` says, as the v1 login client did. The SDK flows (BROWSER, DEVICE, CERTIFICATE) honor `--allowInsecure`.
- **Proposed**: verify certificates unless `--allowInsecure` is set.
- **Breaks**: users with self-signed tenants who never set `--allowInsecure`.
- **Ways to ship**: warn when a login succeeds only because verification was skipped, then flip the default.

### 8. Default page size

- **Today**: the CLI always sends `pageSize=20` unless the setting is the empty string. The SDK default is 5, and an explicit 5 is treated as unset.
- **Proposed**: send no `pageSize` unless asked (the platform default applies), and fix the SDK so an explicit 5 is sent.
- **Breaks**: scripts that rely on getting 20 items per page by default.

### 9. JSON layout opt-out

- **Today (v3)**: pretty JSON prints short objects inline (`{ "pageSize": ["1"] }`), a change from v2's layout. The JSON is the same; text diffs differ.
- **Proposed**: keep the new layout and document it, or add an option for the old one.
- **Breaks**: golden-file comparisons of pretty output.

### 10. Extension SDK version

- **Today**: `pkg/extsdk` v0.0.1 was tagged with a v1 client. On v3, `extsdk.Context.Client()` returns the go-c8y v2 `*api.Client`.
- **Decision**: the next extsdk tag (v0.x allows breaking changes, but extension authors need a migration note: v1 `c8y.Client` to v2 `api.Client`).

### 11. Go SDK API cleanups (breaking for SDK callers if done in place)

These come from the SDK consistency audit. Each can also be done additively (new name plus deprecation) and removed later.

| Topic | Today | Proposed |
|---|---|---|
| Id types | `userroles/inventory` takes `int64` ids, operations `BulkOperationID int`; everything else uses strings | string ids everywhere |
| Option names | singular `GetOption`, `UpdateOption`, `TenantOption`, ... next to `...Options` | `...Options` everywhere |
| Create/Update shapes | eight or more signatures (`Create(ctx, body)`, `Create(ctx, opts)`, `Create(ctx, parentID, body)`, ...) | `Create(ctx, CreateOptions)` with an optional raw `Body` |
| `ResolveID` | five signatures | one signature |
| `NewClient` | no error return; a bad URL is silently ignored | `NewClient` returns an error (or keep it and add `NewClientE`) |
| Filter fields | `Type []string` in alarms, `string` in events; `Source DeviceRef` versus `DeviceID`/`AgentID` | one field type per concept |
| Sub-services | three `Microservices` sub-services are struct values, the rest pointers | pointers everywhere |
| Names | duplicate type names (`GroupRef`), empty `Ref` stubs, per-package iterator aliases, `Client.Configuration` versus `Repository.Configuration` | one name per concept |
| `SendRequest` | returns the v1-style `*RequestResult` | `Client.Request` (added in #225) as the only raw API |
| `Result.First()` | error on empty, while `op.First` returns a bool | one convention |
| Method names | `binaries.Get` downloads, `events.ListV2`, `...All` names | `Download`, one `List`, aligned names |

### 12. Removing deprecated SDK aliases

Deprecated in go-c8y v2 and kept as aliases: `Client.ApplicationVersions`, `DeviceStatistics`, `LoginTokens`, `DevicePermissions`, `InventoryRoles`, `Users.CurrentUser`, `Microservices.CurrentMicroservice`, `Devices.CreateDevice`, `Agents.CreateAgent`, `realtime.NewClient`, `notification2.NewNotification2Client`, `SetMicroserviceHost`, `encrypt.Encrypt*`, `NewDefaultMicroservice`, `SubscribeToNotifications`, `SubscribeToOperations`.

- **Decision**: the deprecation window (for example remove in v2.x+2 or only in v3).

### 13. `devices enroll` fails fast on permanent errors

- **Today**: `devices enroll` (and the SDK's `devices.Enroll`) retries every error, including permanent 4xx answers from the EST endpoint, until the timeout is reached.
- **Proposed**: stop at once on permanent errors (for example 400, 401, 403, 404) and keep retrying only on network errors, 5xx and "not accepted yet" answers.
- **Breaks**: scripts that rely on the retries to bridge a short window where the endpoint answers 4xx while the device request is still being set up; the command would then fail sooner.
- **Ways to ship**: opt-in first (a `--failFast` flag), then flip the default in a major release.
- **Source**: go-c8y `devices.Enroll`, `c8y devices enroll`.

### 14. `devices enroll --mode manual` keeps an existing device request

- **Today**: with session credentials, `--mode manual` deletes an existing device request with the same external id when it starts, so a request that a user registered beforehand is lost.
- **Proposed**: do not remove the request in manual mode; use it, and only fail when the request is in an unusable state.
- **Breaks**: scripts that rely on the clean start (an old, stale request is no longer replaced).
- **Ways to ship**: opt-in first (a flag to replace the request), then flip the default in a major release.
- **Source**: `c8y devices enroll --mode manual`.

### 15. Plain-number timeout values

- **Today**: a plain number in `C8Y_SETTINGS_DEFAULTS_TIMEOUT` (or `settings.defaults.timeout`) is read as seconds, the same as v2. A value written in milliseconds therefore means hours, and `devices enroll` waits that long by default.
- **Proposed**: warn on large plain numbers (for example above one day) or require a unit (`30s`, `5m`).
- **Breaks**: with a required unit, every setting that holds a plain number; with only a warning, nothing.
- **Ways to ship**: warning first, require the unit in a major release.
- **Source**: settings handling for `defaults.timeout`.

### 16. `--raw` output of list commands that return a bare array

- **Today**: `features list` (new in v3, bare-array response) prints one object per item under `--raw`; there is no v2 reference, so the raw output was left as is. `ui applications plugins list` (new in v3) prints the response envelope under `--raw`, like the other list commands.
- **Proposed**: decide one rule for new commands: either `--raw` prints the response as received (envelope or bare array) everywhere, or it prints the items everywhere.
- **Breaks**: only users of the two new commands.
- **Ways to ship**: change before the first v3 release, when nobody relies on it yet.
- **Source**: CLI#820.

### 17. `c8y api` with an absolute url

- **Today**: a piped absolute url (for example `https://test.com/path`, or a self link from a response) is reduced to its path and query, and the request goes to the session's host (or `--host`), the same as v2. The host of the url is dropped without a message, so `echo https://test.com/path | c8y api` calls the session's tenant, not test.com.
- **Proposed**: honor the host of an absolute url, and send the session's credentials only when it is the session's own host (never to another host). Alternatively, fail with a clear error when the host differs from the session's.
- **Breaks**: scripts that pipe self links from another tenant or an old host name and rely on them being redirected to the session's host (honoring the host would call the other host; failing would be an error).
- **Ways to ship**: an opt-in flag first (for example `--followHost`), or a warning when the host differs, then the stricter behavior in a major release.
- **Source**: `pkg/cmd/api` `requestURL`.

### 18. Create commands are not idempotent

- **Today**: `ui plugins create` and `microservices serviceusers create` (and similar create commands, such as creating a plugin version or an application that exists) fail with a 409 when the version or application already exists, so a script cannot be re-run.
- **Proposed**: an opt-in, such as `--ifExists skip` (or `update`), or treating an identical existing version (same name and content) as success. This is the per-command part of topics 2 and 3.
- **Breaks**: nothing if opt-in. Treating identical versions as success changes the exit code from 9 to 0 for scripts that use the 409 as a signal.
- **Ways to ship**: `--ifExists` on the create commands first, default `fail`.
- **Source**: live test runs of `ui plugins create` and `microservices serviceusers create`.

### 19. `c8y api` and the paging flags

- **Today**: for a GET, `c8y api` sends only the query of the url and `--customQueryParam` (the same as v2). `--pageSize`, `--currentPage`, `--withTotalPages` and `--withTotalElements` are accepted but not sent, and the default page size is not sent either, so the server's default applies. `--includeAll` and `--totalPages` still follow the `next` links. The help examples `c8y api GET "/alarm/alarms&status=ACTIVE" --pageSize 10` and `... --pageSize 1 --withTotalPages` therefore do not do what they say (v3 only: v2 has no `--withTotalElements`, and rejects it as an unknown flag).
- **Proposed**: send the paging flags the user sets explicitly (not the defaults), unless the url already has the parameter, so `c8y api ... --pageSize 10` works like on the list commands. Alternatively, remove the flags and the examples from `c8y api`.
- **Breaks**: scripts that pass these flags to `c8y api` today get changed requests (a smaller or larger page, or `withTotalPages=true` in the response). Binary downloads and other non-collection paths must not receive them, so the flags would apply to the requests that have them set only.
- **Ways to ship**: honor explicit flags in a major release, with a release note. The v3 branch briefly sent `pageSize` (the default) on every GET, including binary downloads; that was a regression and is fixed.
- **Source**: this page's measurements of v2 (master) and v3 against a stub server.

## Part 2: already changed on v3 (confirm or revert)

These changes are on the v3 branch today. A v2 CLI user would notice them. Each should be confirmed (and listed in the release notes) or reverted.

### Exit codes

| Change | v2 CLI | v3 | Source |
|---|---|---|---|
| Extension command missing a required pipeline path parameter | 100 | 101 | CLI#801 |
| `ui plugins create` errors (not a plugin, file missing) | 100 | 101 | CLI#805 |
| `events list --device <file outside the allowed paths>` | 101 | 100 (resolved like `--id`) | CLI#806, `tests/manual/fileguard/fileguard.yaml` |
| `operations wait` on a missing id | 100, "No operation for" | 4, message contains "404" | `tests/manual/operations/wait/operations_wait.yaml` |
| Name that cannot be resolved (built-in commands) | 100 (unverified) | 101 (unverified) | CLI#801, CLI#805 |
| `microservices delete --id <missing name>` | 4 | 4 (same), but the message is now `microservice not found with name: <name> [404]` | CLI#811 |

### Behavior

- **Dry run is more complete**: lookups always run. `ui plugins create` downloads a URL under `--dry` (v1 failed). An extension `binaryUploadURL` parameter no longer uploads for real under `--dry` (v1 did) and shows a placeholder URL. `serviceusers create --dry` shows the create and subscription requests (CLI#805).
- **Extension commands**: `user[]`/`userself[]` no longer expand wildcards; `hostedapplication` resolves any application by name (v1: only hosted applications of the current tenant); piped items whose lookup fails now produce item errors instead of being skipped (CLI#801).
- **`agents list`** sends the query in a different order (`type eq 'type1' and (has(com_cumulocity_model_Agent))` versus `(has(Agent)) and type eq 'type1'`); same result (CLI#806).
- **`inventory find`** may send `query=... and has(c8y_IsDevice)` where v2 sent `q=...` (seen only in a test change; needs verification).
- **Multi-value flags**: `microservices serviceusers get --id` and `remoteaccess configurations update --device` accept lists (CLI#799, CLI#800).
- **`sessions login`**: `--tfaCode` is now used (it was ignored); `sessions set --tfaCode` is a code, not a TOTP secret (CLI#804).
- **`--allowInsecure` and proxy settings** now apply to every command (before, only through the legacy client) (CLI#804).
- **Cache**: `C8Y_SETTINGS_CACHE_METHODS` is supported; responses with status 300 or above are not cached.
- **Assert commands**: `--strict` with several ids prints one line per id plus a summary (three lines instead of two); the timeout message lost its ", context" suffix; `--not` on an existing object waits the full duration.
- **`c8y api`**: HTTP errors become item errors with the right exit code; HTML error bodies are shown; `--withTotalPages`, `--withTotalElements` and `--raw` print the whole response; a URL with its own `pageSize` no longer gets a second one (CLI#797).
- **Logging**: notification2 SDK logs go to stderr at warning level (more with `--verbose`/`--debug`); remote access logs only with `--verbose`.
- **Session handling**: the tenant is taken from the token when possible, so the first command of a session no longer makes an extra request (CLI#804).

### Output format

- Pretty JSON prints short objects inline (see topic 9).
- Markdown table separator `|---|` became `|:---|`; arrays of objects get one column per index; selected array elements get their own column; SI numbers lose a trailing space.
- Error text follows the SDK error format (status, type, message); HTML error bodies are replaced by the status text.
- `--verbose` request lines include paging parameters; `--dry` output includes `User-Agent`, `X-Application` and `Accept-Encoding` headers; multipart parts carry a charset.
- The success confirmation for commands without a response body (v1 "Updated ... => 200 OK") may not be ported (needs verification).

### Extension SDK and Go packages

- `extsdk.Context.Client()` returns the v2 client (topic 10).
- Deleted CLI packages that third-party Go code might import: `pkg/request`, `pkg/requestiterator`, `pkg/worker`, `pkg/c8yfetcher`, `pkg/c8yquerycmd`, `pkg/c8ywaiter`, `pkg/desiredstate`, `Factory.Client` and the `RunWithWorkers` family; command packages merged from 359 to 87 (CLI#777, CLI#804, CLI#805, CLI#807).

### Go SDK behavior (go-c8y v2, relative to earlier v2 pre-releases)

- Deletes, unassigns and unsubscribes ignore a 404 by default (`api.WithNotFoundAsError` opts out) (#234).
- Tokens are fetched lazily; `SetAuth` replaces credentials; `AuthType` applies per request; RSA client certificates use the HTTPS exchange by default (#233).
- A pure dry run returns status Skipped; `IsSuccess()` is true for no-content and skipped results (#235).
- Websocket subscriptions do not connect under dry run (#230).
- Paging stops with a "pagination made no progress" error instead of hanging (#232).
- `agents.ListOptions.Query` is sent as `q` (#235).
- `SendRequest` reports HTTP error statuses as errors (#220).
- `api.WithNotFoundAsError` and `api.WithConflictAsError` only change the SDK's own `IgnoreNotFound`/`IgnoreConflict` calls. A caller's explicit `.IgnoreStatusCodes(404)` still clears the error, which is intended (the caller asked for it) but worth documenting next to the options.
- `Microservices.Delete` with a name that does not resolve returns Skipped ("not found") even with `api.WithNotFoundAsError`, because the not-found comes from the name lookup, not from a 404 on the delete. The CLI resolves the name itself to keep exit 4 (CLI#811). Making the SDK honor the option here is a fix for callers who set it, and a change for callers who rely on Skipped.
- `Registration.Register` on an existing external id used to fail against the fake server ("already exists"). With a new password it now succeeds and updates the credentials, like the platform (the response lists the id in `credentialUpdatedList`); with the same password the row still fails ("Error during password validation."). Callers that relied on the error must check `credentialUpdatedList` instead. The CLI test `tests/manual/deviceregistration/register_bulk.yaml` covers this.
- The fake server's 404 error type changed from `<resource>/notFound` to the platform style `<resource>/Not Found` for managed objects, alarms, events, measurements and operations. Code that matches the old type string breaks.

## Part 3: still to verify

Verified since the first version of this page: a 409 on identity create, child assign or subscribe exits 9 as in v2 (CLI#809, `tests/manual/common/errors/conflict_not_ignored.yaml`).

- The exit code for unresolved names on built-in commands (100 or 101).
- The `inventory find` query shape, and whether the no-content success message is ported.
- `--withGroups` is accepted but ignored by `devices get`, `devices list`, `agents get`, `agents list` and `smartgroups get`, although the SDK now has `managedobjects.GetOptions.WithGroups` (other inventory-based commands already send it). Wiring it would add `withGroups=true` to the request and `assetParents` to the output when the flag is set, so it changes output for scripts that pass the flag today.

## Notes and open questions (not CLI behavior)

- **Live tests that need a tenant feature**: a live test cannot skip itself when the tenant lacks a feature. For example, Notification 2.0 answers the test user with 403 (`security/Forbidden`), so the test fails although the CLI works. This is a convention for the test suite, not a CLI change. Proposed: a shared helper in `tests/manual` that probes the feature once (a cheap request, a 403 or 404 meaning "not available") and marks the test as skipped, and a list of the features that the shared test tenant has. Decision needed: where the helper lives (a script sourced by the tests, or a commander precondition) and whether a skipped test counts as a pass in the live workflow summary.
- **`--includeAll` of `c8y api` against a stub**: with a stub that returned a next link on pages 1 and 2 (three pages in all), the v2 binary stopped after the second page, while v3 follows every page. To verify against a real tenant whether v2 had a limit (it is not documented), and keep v3's behavior if not.
