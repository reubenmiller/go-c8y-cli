# Breaking-change candidates (go-c8y-cli v3 and go-c8y v2)

This page collects improvements that would change existing behavior. Each one needs a decision before it ships, because it can break scripts, extensions or SDK callers.

The reference is the released v2 CLI: tag `v2.59.0` on branch `v2`. (Branch `master` is a stale 2022 branch and is not a reference.)

Rule for v3: keep or restore the `v2.59.0` behavior. A better but breaking idea is not implemented. It is listed here until it is decided. Opt-in additions, such as a new flag or a new command that does not change any default, can ship without a decision.

Sources: go-c8y-cli PRs #777 to #831, go-c8y PRs #220 to #245, the go-c8y v2 CHANGELOG, the SDK consistency audit, and a side-by-side run of a `v2.59.0` build and a build of `v3` (commit `59de5b0d6`) against the go-c8y fake server. Part 2 lists only differences that still exist in that run.

## How to read an entry

- **Today**: what users get now (`v2.59.0`, and v3 where it differs).
- **Proposed**: the better behavior.
- **Breaks**: who notices.
- **Ways to ship**: options that limit the damage, usually one or more of:
  1. opt-in first (flag, setting or environment variable), default unchanged;
  2. deprecation warning when the old behavior is relied on;
  3. flip the default in a major release, with the old behavior still available as an opt-out.

## Notes before the lists

### Restored on v3 (same as v2.59.0, not a break)

Earlier versions of this page recorded some v2 behavior as v3 changes. These are not breaks:

- The audit of `v2.59.0` showed that pretty JSON layout, the markdown table separator `|:---|`, SI number format, the `agents list` query order, the `inventory find` `query=` parameter, `C8Y_SETTINGS_CACHE_METHODS` and the no-cache rule for status 300 and above, the assert timeout text and the "Updated ... => 200 OK" confirmation were already in v2.59.0. They were removed from this page.
- These v3 differences were found by the audit and fixed (go-c8y-cli #827 to #831, go-c8y #243 to #245):
  - error text, error type and exit codes of built-in commands, including the `Error Response` block, the `--withError` document, assert commands and the denied `--device` file rule;
  - request shape: one-page lists send the plain request, `--retries` with `retryWaitMin` and `retryWaitMax`, `settings.http.compression`, `Accept` and the `go-client` identity headers, multipart file part without a charset, no `queryParams` field in dry JSON, `--raw` with `--select`, `--flatten` and `--filter` on a top-level array, `c8y api` with an absolute url;
  - `--verbose` log lines, the characters `<`, `>` and `&` in output, custom `User-Agent` and `X-Application` headers, retry backoff without jitter;
  - realtime and notification2 websockets over TLS, `--dry` for websocket commands, the remote access "configuration not found" text.

### Follows the v2 release (not a break)

v3 had not yet taken these v2.59.0 changes. They were merged into v3 (go-c8y-cli #827):

- `devicegroups create --type` accepts `c8y_DeviceSubgroup` (v2 #779). A script that used `c8y_DeviceSubGroup` is affected, as it is on v2.59.0.
- `c8y mcp serve --mode` is the ceiling, the highest mode the agent can reach (default `dev`). The user can raise the session mode up to it (v2 #812).
- `c8y mcp serve --sessionProvider` is new (v2 #781). The MCP server rejects unknown sub commands (#780), handles concurrent tool calls (#795) and classifies `applications open --noBrowser` as a read (#774).

### Restored behavior that v3 still differs from in one detail

- `--debug` prints the v2.59.0 lines again except `timeout:`, `checking job iterator`, `Input line`, `finished adding jobs` and `Overriding common options` (see Part 2).

## Part 1: proposed improvements (decision needed)

### 1. Idempotent delete (missing id is not an error)

- **Today**: deleting an id that does not exist exits 4 (one item) or 104 (several items), on `v2.59.0` and on v3. Scripts silence it with `--silentStatusCodes 404` and `--silentExit`. The go-c8y v2 SDK treats a 404 on delete, unassign and unsubscribe as success (`IgnoreNotFound`); the CLI turns it back into an error with `api.WithNotFoundAsError` so v3 keeps the v2 exit codes.
- **Proposed**: "make sure it is gone" semantics. A missing id is success, as in the SDK, Terraform and most declarative tools. Cleanup scripts need no extra flags.
- **Breaks**: scripts that check exit 4/104 after a delete, or use the 404 as a signal that the object was already removed.
- **Ways to ship**: an opt-in setting (for example `settings.defaults.idempotentDelete` or a `--ignoreNotFound` flag), then a later default flip with `--failOnNotFound` as the opt-out. `--silentStatusCodes 404` keeps working either way.
- **Source**: CLI#806, go-c8y#234, `tests/manual/common/errors/test_delete.sh`.

### 2. Idempotent create and subscribe (409 conflict is not an error)

- **Today**: creating something that already exists (external id, child assignment, application subscription, notification2 subscription) fails with exit 9, on `v2.59.0` and on v3 (CLI#809, `tests/manual/common/errors/conflict_not_ignored.yaml`). The SDK treats these as success (`IgnoreConflict`, status Duplicate); the CLI turns it back into an error with `api.WithConflictAsError`.
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

### 4. Websocket commands: dry run, TLS and proxy

Websocket commands are: `alarms|events|measurements|operations|inventory subscribe`, `realtime subscribe`, `realtime subscribeAll`, `notification2 subscriptions subscribe` and the remote access commands (`remoteaccess server`, `connect ssh`, `connect run`). On v3 they behave as in `v2.59.0` (go-c8y-cli #831). Four things could be better and would break something.

**4a. `--dry` does not connect**

- **Today**: the subscribe commands (including `realtime subscribeAll`) connect and stream for the full `--duration` under `--dry`. `notification2 subscriptions subscribe --dry` also shows the two token requests as What If. `remoteaccess server` and `connect ... --dry` start the local listener and do not dial. The SDK itself does not connect under dry run (go-c8y#230); the CLI connector dials with the dry-run flag removed from the context.
- **Proposed**: under `--dry`, print the upgrade request and the channels, and do not connect.
- **Breaks**: anyone using `--dry` with these commands expecting a live stream.

**4b. Realtime honors the insecure setting**

- **Today**: realtime websockets never verify the certificate, whatever `settings.defaults.insecure` says, as in `v2.59.0`. `notification2 subscriptions subscribe` honors the setting.
- **Proposed**: realtime honors the setting like every REST command (a security improvement).
- **Breaks**: users who rely on self-signed edge certificates without `--allowInsecure`.
- **Ways to ship**: warn when verification was skipped, then flip the default.

**4c. Remote access honors the insecure setting**

- **Today**: `remoteaccess server`, `connect ssh` and `connect run` verify certificates and ignore the setting (as in `v2.59.0`). A self-signed edge cannot be reached through remote access even with `--allowInsecure`, while REST works.
- **Proposed**: honor the setting.
- **Breaks**: nothing for users with trusted certificates. The change only makes more setups work.

**4d. TLS and proxy settings come from one place**

- **Today**: realtime honors `--ignoreProxy`; notification2 and remote access always use the environment proxy. The CLI sets TLS and proxy for websockets in `pkg/cmdutil/wsconnector.go`, because the SDK builds its websocket dialer from the REST transport and the CLI wraps that transport, so the SDK cannot see through it. The realtime dialer also uses the SDK default handshake timeout of 45 seconds (`v2.59.0`: 10 seconds), and pressing Ctrl-C during the TCP or TLS handshake waits for that timeout.
- **Proposed**: one rule for TLS, proxy and handshake timeout in all websocket commands. SDK side: an explicit websocket TLS and proxy option on `ClientOptions`, and an exported cancel-aware dial helper, so the CLI can drop the connector.
- **Breaks**: users who depend on the current proxy split (for example `--ignoreProxy` affecting only realtime).

### 5. One exit-code table and error format

- **Today**: 100 (unexpected error) and 101 (user error) are used inconsistently between built-in, extension and runtime commands. v3 now reproduces the `v2.59.0` codes and error text for built-in commands, including the inconsistencies below.
- **Proposed**: a single mapping (HTTP status codes, user errors, unresolved names, timeouts, assertion failures), documented and tested, with one error format. Known inconsistencies to clean up:
  - a denied `--device` file exits 101 for filter and body flags and 100 for path arguments and assert commands;
  - name lookup failures print `commandError: referenceByName: no matching items found ... ::StatusCode=404`, while a direct lookup (`usergroups getByName`) prints a `serverError` with the Error Response block; the SDK text (`managed object not found with name: x`) is clearer;
  - the remote access "not found" error has three texts and two exit codes: `configuration not found. name=` (101) for connect and server, `referenceByName: no matching items found` (4) for `configurations get` and `delete`; v3 uses a fourth text (101) for `get` and `delete` (see Part 2);
  - REST transport errors (connection refused, TLS verification failure) exit 99 with the raw Go message on `v2.59.0`;
  - an error whose message equals the server message repeats the text after `METHOD url: status type`;
  - `--withError` prints the error document twice on `v2.59.0` (sorted and nested);
  - the aggregate "jobs completed with N errors" line depends on the worker type, not on the command kind;
  - `operations wait --id 1,2` on `v2.59.0` reports only the last error and exits 100;
  - `microservices serviceusers get --id <unknown name>` exits 0 with no output on `v2.59.0`;
  - error text for a response without a JSON body: v1 printed an empty message, v3 derives it from the SDK;
  - the v1 `details` field of the error JSON (debug mode only) is not restored.
- **Breaks**: scripts that match the exit codes or messages that change.
- **Source**: CLI#828, CLI#831.

### 6. Name resolution: wildcards and ambiguous matches

- **Today (v3)**: user lookups in extension commands no longer expand wildcards (`--user 'adm*'`; `v2.59.0` expands them and takes the first match). The SDK's new `Users.ResolveID` supports wildcards but returns an "ambiguous" error when more than one user matches. Role lookups keep wildcard and first-match.
- **Proposed**: one rule for all name lookups: wildcards allowed; more than one match is an error that lists the candidates (safer for update and delete), with an opt-in `--first` for the old behavior.
- **Breaks**: scripts that rely on the first match.

### 7. Login TLS verification

- **Today (kept on purpose)**: BASIC, OAUTH2_INTERNAL and OAUTH2 login and the login-options lookup accept self-signed certificates whatever `--allowInsecure` says, as the v1 login client did. The SDK flows (BROWSER, DEVICE, CERTIFICATE) honor `--allowInsecure`.
- **Proposed**: verify certificates unless `--allowInsecure` is set.
- **Breaks**: users with self-signed tenants who never set `--allowInsecure`.
- **Ways to ship**: warn when a login succeeds only because verification was skipped, then flip the default.

### 8. Default page size and the stray `pageSize`

- **Today**: the CLI sends `pageSize=20` unless the setting is blank (a space) or `0`; an empty string still sends 20. An explicit `5` (setting or `--pageSize 5`) sends no `pageSize` at all, because of a quirk inherited from the v1 SDK (the SDK default is 5). Get-by-id requests (`devices get`, `inventory get`) carry `?pageSize=20` on `v2.59.0`; v3 does not send it.
- **Proposed**: send no `pageSize` unless asked (the platform default applies, and nothing is added to get-by-id requests), and send an explicit 5.
- **Breaks**: scripts that rely on getting 20 items per page by default, or on the exact request URL. Sending an explicit 5 is itself a change.

### 9. Extension SDK version

- **Today**: `pkg/extsdk` v0.0.1 was tagged with a v1 client. On v3, `extsdk.Context.Client()` returns the go-c8y v2 `*api.Client`.
- **Decision**: the next extsdk tag (v0.x allows breaking changes, but extension authors need a migration note: v1 `c8y.Client` to v2 `api.Client`).

### 10. Go SDK API cleanups (breaking for SDK callers if done in place)

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
| Identity headers | `middlewareAddUserAgent` has its arguments swapped: `User-Agent` always gets `go-client` and `X-Application` gets `ClientOptions.Agent`; the CLI passes `go-client` for both, so nothing shows today | fix the swap (changes the headers of SDK callers) |
| Websocket TLS and proxy | the dialer reads TLS and proxy from the REST transport (see topic 4d) | an explicit option on `ClientOptions` |

### 11. Removing deprecated SDK aliases

Deprecated in go-c8y v2 and kept as aliases: `Client.ApplicationVersions`, `DeviceStatistics`, `LoginTokens`, `DevicePermissions`, `InventoryRoles`, `Users.CurrentUser`, `Microservices.CurrentMicroservice`, `Devices.CreateDevice`, `Agents.CreateAgent`, `realtime.NewClient`, `notification2.NewNotification2Client`, `SetMicroserviceHost`, `encrypt.Encrypt*`, `NewDefaultMicroservice`, `SubscribeToNotifications`, `SubscribeToOperations`.

- **Decision**: the deprecation window (for example remove in v2.x+2 or only in v3).

### 12. `devices enroll` fails fast on permanent errors

- **Today**: `devices enroll` (and the SDK's `devices.Enroll`) retries every error, including permanent 4xx answers from the EST endpoint, until the timeout is reached (`v2.59.0` does the same).
- **Proposed**: stop at once on permanent errors (for example 400, 401, 403, 404) and keep retrying only on network errors, 5xx and "not accepted yet" answers.
- **Breaks**: scripts that rely on the retries to bridge a short window where the endpoint answers 4xx while the device request is still being set up; the command would then fail sooner.
- **Ways to ship**: opt-in first (a `--failFast` flag), then flip the default in a major release.
- **Source**: go-c8y `devices.Enroll`, `c8y devices enroll`.

### 13. `devices enroll --mode manual` keeps an existing device request

- **Today**: with session credentials, enroll deletes an existing device request with the same external id when it starts, in any mode, so a request that a user registered beforehand is lost. (`v2.59.0` does the same.)
- **Proposed**: do not remove the request in manual mode; use it, and only fail when the request is in an unusable state.
- **Breaks**: scripts that rely on the clean start (an old, stale request is no longer replaced).
- **Ways to ship**: opt-in first (a flag to replace the request), then flip the default in a major release.
- **Source**: `c8y devices enroll --mode manual`.

### 14. Plain-number timeout values

- **Today**: a plain number in `C8Y_SETTINGS_DEFAULTS_TIMEOUT` (or `settings.defaults.timeout`) is read as seconds, the same as `v2.59.0`. A value written in milliseconds therefore means hours, and `devices enroll` waits that long by default.
- **Proposed**: warn on large plain numbers (for example above one day) or require a unit (`30s`, `5m`).
- **Breaks**: with a required unit, every setting that holds a plain number; with only a warning, nothing.
- **Ways to ship**: warning first, require the unit in a major release.
- **Source**: settings handling for `defaults.timeout`.

### 15. `--raw` output of list commands that return a bare array

- **Today**: `features list` (new in v3, bare-array response) prints one object per item under `--raw`; there is no v2 reference, so the raw output was left as is. `ui applications plugins list` (new in v3) prints the response envelope under `--raw`, like the other list commands.
- **Proposed**: decide one rule for new commands: either `--raw` prints the response as received (envelope or bare array) everywhere, or it prints the items everywhere.
- **Breaks**: only users of the two new commands.
- **Ways to ship**: change before the first v3 release, when nobody relies on it yet.
- **Source**: CLI#820.

### 16. `c8y api` with an absolute url

- **Today**: on `v2.59.0` and on v3, `c8y api` requests an absolute url (argument or piped, for example a self link from a response) from the url's own host, with the session's credentials, and logs the warning "Parsing url in request". Credentials therefore go to any host that a url names. Two details differ, and v3 keeps its behavior: `v2.59.0` showed the session's host in the `Host` line of a dry run (v3 shows the url's host), and with `--host <host>/<prefix>` `v2.59.0` built a broken path on the `--host` host (`/<prefix>/http:/other/...`) where v3 requests the url's own host.
- **Proposed**: send the session's credentials only when the url is on the session's own host (never to another host), or fail with a clear error when the host differs.
- **Breaks**: scripts that call another host through an absolute url and rely on the session's credentials being sent there (they would get no credentials, or an error).
- **Ways to ship**: a warning when the host differs, then the stricter behavior in a major release.
- **Source**: `pkg/cmd/api` `requestURL`; CLI#824, CLI#829.

### 17. Create commands are not idempotent

- **Today**: `ui plugins create` and `microservices serviceusers create` (and similar create commands, such as creating a plugin version or an application that exists) fail with a 409 when the version or application already exists, so a script cannot be re-run.
- **Proposed**: an opt-in, such as `--ifExists skip` (or `update`), or treating an identical existing version (same name and content) as success. This is the per-command part of topics 2 and 3.
- **Breaks**: nothing if opt-in. Treating identical versions as success changes the exit code from 9 to 0 for scripts that use the 409 as a signal.
- **Ways to ship**: `--ifExists` on the create commands first, default `fail`.
- **Source**: live test runs of `ui plugins create` and `microservices serviceusers create`; CLI#824.

### 18. Logging and output rendering

- **Today**: `--verbose` and `--debug` print the `v2.59.0` line format (no HTTP dump, no SDK log lines). Output replaces the escaped characters `<`, `>` and `&` with `<`, `>` and `&` in all output formats, with a byte-level replace over the whole text (an escaped backslash followed by `u003c` is also altered on `v2.59.0`; v3 replaces only the real escapes in that case so the JSON stays valid). Under `--workers N`, the worker number in `--verbose` is the SDK pool slot, not the v2 worker id.
- **Proposed**: structured (slog) logging with the request and response dump behind a separate flag (v3 had this before #830); configure the JSON encoder not to escape HTML in the first place and never rewrite output text; number workers by a real worker id.
- **Breaks**: scripts that parse `--verbose` lines, and anyone who relies on `<`, `>` and `&` in output (they would see escapes again unless the encoder change is made together).
- **Source**: CLI#830.

### 19. Retry backoff with jitter

- **Today**: `--retries N` waits `retryWaitMin`, then twice as long each time, up to `retryWaitMax`, without jitter (`RetryOptions.Exponential`, go-c8y#245), as in `v2.59.0`. Many parallel clients retry in lockstep. A `Retry-After` header on a 429 or 503 response takes precedence.
- **Proposed**: jittered backoff by default (v3 did this before #830).
- **Breaks**: scripts or tests that depend on exact retry timing.
- **Ways to ship**: opt-in setting first, then flip the default.
- **Source**: CLI#830, go-c8y#245.

### 20. Paging strategy for `--includeAll` and `--totalPages`

- **Today**: for `--totalPages N` with N above 1, the first request of `devices list` differs from `v2.59.0`: v2 replaced the query with the id keyset (`_id gt '0'`, `$orderby=_id asc`), v3 keeps `$orderby=name` and sends `currentPage=1` (reported for `--includeAll` too: the SDK uses the id keyset for inventory and `currentPage=1` on the first request). `--paginationStrategy` (auto, offset, id) exists only on v3.
- **Proposed**: decide whether the SDK strategy is the rule everywhere, and document the order and the keyset.
- **Breaks**: scripts that depend on the order of results with `--includeAll`.

### 21. Object type of `binaries create` for a file without extension

- **Today**: the object type is `application/octet-stream` on `v2.59.0`; v3 gets `text/plain; charset=utf-8` because the SDK sniffs the content (measured with a text file named `nofile`).
- **Proposed**: keep the SDK behavior (it follows the content) or restore octet-stream; decide.
- **Breaks**: scripts that match the type of binaries created from extensionless files.

### 22. Per-request timeout model

- **Today**: `--timeout` is restored. A real deadline that starts before the request is prepared (the effect of `--timeout 1ms` on `v2.59.0`, a race even there: 0 to 3 of 10 runs timed out on loopback) is not reproduced, so a timeout test needs a slow server.
- **Proposed**: define when the timeout starts (request preparation, connect, whole command).
- **Breaks**: scripts that depend on very short timeouts.

## Part 2: differences that remain on v3 (confirm or revert)

These differences from `v2.59.0` exist on v3 today (measured on the fake server unless a source is named). A v2 CLI user would notice them. Each should be confirmed (and listed in the release notes) or reverted.

### Exit codes and errors

| Change | v2.59.0 | v3 | Source |
|---|---|---|---|
| Extension command missing a required pipeline path parameter (`kitchensink pipeline get`) | exit 100, `missing required parameters. [device]` | exit 101, `Missing required parameter. device` | CLI#801 |
| `remoteaccess configurations get` and `delete` with an unknown configuration name | exit 4, `referenceByName: no matching items found. name=nope ::StatusCode=404` | exit 101, `remote access configuration not found: name=nope` | CLI#831 |
| `microservices serviceusers get --id <unknown name>` | exit 0, no output | exit 4, `referenceByName: no matching items found` | CLI#799 |
| REST transport error (connection refused, TLS verification failure) | exit 99, `serverError: Get "<url>": <error>` | exit 101, `commandError: [002]  : Get "<url>": <error>` | CLI#828 |
| `operations wait --id 1,2` with several failing ids | exit 100, only the last error | exit 104, one error per id and `jobs completed with 2 errors` | CLI#828 |
| `--withError` with a failing request | prints the error document twice | prints it once | CLI#828 |

### Behavior

- **Dry run is more complete**: lookups always run. `ui plugins create --file <url>` downloads the url under `--dry` (v2 failed with "no such file"). An extension `binaryUploadURL` parameter no longer uploads for real under `--dry` (v2 did) and shows a placeholder url in the body (CLI#805).
- **Extension commands**: `user[]` and `userself[]` no longer expand wildcards (`--user 'peter*'` is sent as written; v2 resolved it to the first match). `hostedapplication` resolves any application by name (v2: only hosted applications of the current tenant, and `reference by name: no matches found` for a microservice name) (CLI#801).
- **`--withGroups` is ignored** by `devices get`, `devices list`, `agents get`, `agents list` and `smartgroups get`. `v2.59.0` adds `withGroups=true` to the request. (`inventory list` and `devicegroups get` still send it.) The SDK has `managedobjects.GetOptions.WithGroups`, so this looks like a regression and not a decision. A fix restores the v2 request and adds `assetParents` to the output when the flag is set.
- **`--withTotalPages` and `--withTotalElements` with `--select`**: v2 returned the whole response, with the statistics, and `--select` paths start at its top level. v3 prints each item on its own line, so the statistics are lost and `--select statistics.totalElements` matches nothing. This is deferred on purpose: the go-c8y v2 client has to keep the statistics with the results first, and the CLI has no workaround. `c8y api` is not affected.
- **Multi-value flags**: `microservices serviceusers get --id` accepts a list (CLI#799).
- **`c8y api` with its own `pageSize`**: for `c8y api /inventory/managedObjects?pageSize=7&type=a`, `v2.59.0` sends `pageSize=20` (the default page size, or `--pageSize`, overrides the one in the url), while v3 keeps the `pageSize=7` of the url. With `--pageSize 3`, v2 sends 3 and v3 still sends 7. A piped absolute url with its own `pageSize` behaves the same way (see topic 8 for the get-by-id case).
- **Get-by-id requests** carry no `pageSize` on v3 (`v2.59.0`: `?pageSize=20`). The `--verbose` line and the error message show the request that was sent (topic 8).
- **`--totalPages 2` on `devices list`**: the first request keeps the `$orderby=name` query on v3 (`v2.59.0` used the id keyset); see topic 20.
- **`binaries create`** of a file without extension sets the type `text/plain; charset=utf-8` (v2: `application/octet-stream`; topic 21).
- **`--outputTemplate`**: the template runs for each item on v3. `v2.59.0` runs it on the whole array, so `output + {...}` fails there with "Unexpected type object, expected array".
- **CSV output**: `devices list -o csv` has no column for an empty object (`c8y_IsDevice`); `v2.59.0` prints it as `{}` (the first column).
- **`--debug`** does not log the `v2.59.0` worker-loop lines (`timeout:`, `checking job iterator`, `Input line`, `finished adding jobs`, `Overriding common options`, `Array properties`), and `Creating c8y client` appears in another order.
- **Logging**: notification2 subscribe sometimes prints `level=WARN msg="Failed to send message" err="use of closed network connection"` on stderr at the end (1 of 5 runs here). The SDK sends the unsubscribe after the connection is closed. `v2.59.0` prints nothing.
- **Websocket handshake**: the realtime handshake timeout is the SDK default of 45 seconds (`v2.59.0`: 10 seconds), and Ctrl-C during the TCP or TLS handshake waits for it (from code review; not timed).

### Extension SDK and Go packages

- `extsdk.Context.Client()` returns the v2 client (topic 9).
- Deleted CLI packages that third-party Go code might import: `pkg/request`, `pkg/requestiterator`, `pkg/worker`, `pkg/c8yfetcher`, `pkg/c8yquerycmd`, `pkg/c8ywaiter`, `pkg/desiredstate`, `Factory.Client` and the `RunWithWorkers` family; command packages merged from 359 to 87 (CLI#777, CLI#804, CLI#805, CLI#807).

### Go SDK behavior (go-c8y v2, relative to the go-c8y v1 client that `v2.59.0` uses and to earlier v2 pre-releases)

- Deletes, unassigns and unsubscribes ignore a 404 by default (`api.WithNotFoundAsError` opts out) (#234). The CLI opts out, so the CLI keeps exit 4.
- Tokens are fetched lazily; `SetAuth` replaces credentials; `AuthType` applies per request; RSA client certificates use the HTTPS exchange by default (#233).
- A pure dry run returns status Skipped; `IsSuccess()` is true for no-content and skipped results (#235).
- Websocket subscriptions do not connect under dry run (#230). The CLI connector overrides this (topic 4a).
- Paging stops with a "pagination made no progress" error instead of hanging (#232).
- `agents.ListOptions.Query` is sent as `q` (#235).
- `SendRequest` reports HTTP error statuses as errors (#220).
- `remoteaccess_configurations.Detect` returns the sentinel errors `ErrPassthroughNotFound` and `ErrNoPassthrough` (#244). The messages are unchanged.
- `RetryOptions.Exponential` and additive identity headers (a custom `User-Agent` or `X-Application` is kept) are new (#245). `PaginationOptions.SinglePage` sends the plain request (#243); its doc says the request is "unchanged", but `PageSize` still goes through `OptimalPageSize` (with `PageSize` 0 it becomes `MaxItems` or 2000, and the CLI relies on this to send `pageSize=20`).
- `api.WithNotFoundAsError` and `api.WithConflictAsError` only change the SDK's own `IgnoreNotFound`/`IgnoreConflict` calls. A caller's explicit `.IgnoreStatusCodes(404)` still clears the error, which is intended (the caller asked for it) but worth documenting next to the options.
- `Microservices.Delete` with a name that does not resolve returns Skipped ("not found") even with `api.WithNotFoundAsError`, because the not-found comes from the name lookup, not from a 404 on the delete. The CLI resolves the name itself to keep exit 4 (CLI#811). Making the SDK honor the option here is a fix for callers who set it, and a change for callers who rely on Skipped.
- `Registration.Register` on an existing external id used to fail against the fake server ("already exists"). With a new password it now succeeds and updates the credentials, like the platform (the response lists the id in `credentialUpdatedList`); with the same password the row still fails ("Error during password validation."). Callers that relied on the error must check `credentialUpdatedList` instead. The CLI test `tests/manual/deviceregistration/register_bulk.yaml` covers this.
- The fake server's 404 error type changed from `<resource>/notFound` to the platform style `<resource>/Not Found` for managed objects, alarms, events, measurements and operations. Code that matches the old type string breaks.

## Part 3: still to verify

- `sessions login --tfaCode` is now used (it was ignored on the old login path), and `sessions set --tfaCode` is a code, not a TOTP secret (CLI#804). Not run against `v2.59.0` here, because it needs a tenant with two-factor authentication.
- The first command of a session no longer makes an extra request, because the tenant is taken from the token when possible (CLI#804). Not counted against `v2.59.0` here.
- The aggregate "jobs completed with N errors" line was compared for assert commands only. Other generic-worker commands (certificate authority, register-bulk, template execute, sessions encrypt and decrypt) print no line on `v2.59.0` and were not checked on v3.
- The fileguard rule (exit 101 or 100) depends on the flag name and command path, and was compared on about 30 commands. Other commands with a `device` driver flag could differ.
- Name lookup detection depends on the SDK text `not found with name` and on the `groupByName` endpoint. A change in the SDK message would silently fall back to a `serverError`; unit tests cover the current text.
- The TLS cases in `tests/offline/realtime/subscribe_parity.yaml` pass without a TLS fake server (they print "skipped" and exit 0). They only mean something under `C8Y_OFFLINE_TLS=1`, which `task test-cli-offline-realtime` sets.
- The retry timing test in `tests/offline/common/output_parity.yaml` checks wall-clock waits and could be flaky on a heavily loaded runner.
- The CLI `go.mod` is pinned to go-c8y `1acd118` (the merge of #245). Re-pin to the newest `origin/v2` commit when another SDK PR merges.

## Notes and open questions (not CLI behavior)

- **Live tests that need a tenant feature**: a live test cannot skip itself when the tenant lacks a feature. For example, Notification 2.0 answers the test user with 403 (`security/Forbidden`), so the test fails although the CLI works. This is a convention for the test suite, not a CLI change. Proposed: a shared helper in `tests/manual` that probes the feature once (a cheap request, a 403 or 404 meaning "not available") and marks the test as skipped, and a list of the features that the shared test tenant has. Decision needed: where the helper lives (a script sourced by the tests, or a commander precondition) and whether a skipped test counts as a pass in the live workflow summary. Source: CLI#824.
- **Offline test suites share a temp folder**: the cache suite uses `$TMPDIR/go-c8y-cli-cache`, so two runs at the same time clash. Use a private `TMPDIR` per run.
- **Generated command docs are stale**: `docs/go-c8y-cli/docs/cli/c8y/*` differs from the output of the generator on v3 (for example the default value text of `-c` and `-M`). A separate regeneration PR is needed.
- **Hand-written commands skip the generator**: `devicegroups create` on v3 is migrated by hand, so a spec change must also be made in `pkg/cmd/*/*_create.go`.
- **Merge v2 into v3 with a merge commit**, not a squash, so `origin/v2` stays an ancestor and later catch-ups stay small.
