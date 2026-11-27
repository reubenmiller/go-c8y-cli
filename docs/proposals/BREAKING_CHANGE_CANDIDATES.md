# Breaking-change candidates (go-c8y-cli v3 and go-c8y v2)

This page collects improvements that would change existing behavior. Each one needs a decision before it ships, because it can break scripts, extensions or SDK callers.

The reference is the released v2 CLI: tag `v2.59.0` on branch `v2`. (Branch `master` is a stale 2022 branch and is not a reference.)

Rule for v3: keep or restore the `v2.59.0` behavior. A better but breaking idea is not implemented. It is listed here until it is decided. Opt-in additions, such as a new flag or a new command that does not change any default, can ship without a decision.

Sources: go-c8y-cli PRs #777 to #842, go-c8y PRs #220 to #249, the go-c8y v2 CHANGELOG, the SDK consistency audit, and a side-by-side run of a `v2.59.0` build and a build of `v3` (commit `9fd137402`) against the go-c8y fake server. Part 2 lists only differences that still exist in that run.

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
  - request shape: one-page lists send the plain request, `--retries` with `retryWaitMin` and `retryWaitMax`, `settings.http.compression`, `Accept` and the `go-client` identity headers, multipart file part without a charset, no `queryParams` field in dry JSON, `--raw` with `--select`, `--flatten` and `--filter` on a top-level array, `measurements list` sends `revert` only when the flag is given, `c8y api` with an absolute url;
  - `--verbose` log lines, the characters `<`, `>` and `&` in output, custom `User-Agent` and `X-Application` headers, retry backoff without jitter;
  - realtime and notification2 websockets over TLS, `--dry` for websocket commands, the remote access "configuration not found" text.
- Batch 3 restored these (go-c8y-cli #835 to #838, go-c8y #246 and #247), each measured against a `v2.59.0` build:
  - views (default columns): every column of a view is a column, an empty cell where a document lacks it, empty objects show as `{}`, views that match on request path or content type work, the view is detected once per command from the first document (as `v2.59.0`), and an unknown `--view` name is reported (#835);
  - errors: a missing required pipeline path parameter of an extension command (exit 100), `remoteaccess configurations get`, `delete` and `update` with an unknown name (exit 4, `referenceByName`), REST transport errors (exit 99, the text twice, `--withError` identical; wait commands exit 100 and assert commands 112), the failed `notification2` token request (exit 100), the denied `--device` file on the create commands of events, alarms, measurements and operations, `ui applications plugins list` without `--application` (exit 0, no output, also with `--raw`), and `user[]`, `userself[]` and `hostedapplication` name matching in extension commands (wildcards expand again, first match; an unknown name prints `reference by name: no matches found`) (#836);
  - requests: `--withGroups` (also `withChildren`, `withChildrenCount` and `skipChildrenNames` on firmware, software and configuration get), `--totalPages N` on events, alarms, measurements, operations and `inventory list` (page by page, as v2), the paging flags and the default page size replace the `pageSize` of the `c8y api` url, the default `pageSize` on get-by-id and similar GET requests (topic 8), `alarms|events|measurements|operations assert count` with `withTotalPages`, `datahub jobs get` (`pageSize`) and `datahub jobs listresults` (`limit`), the plain `file` form field of `register-ca`, and the object type of `binaries create` (the extension of the file, else `application/octet-stream`, no sniffing; `--name` does not change it) (#837);
  - output: `--raw` of `features list` and of every list that returned items before (13 commands) prints the response as received, `ui applications plugins list` prints one document per plugin, `--outputTemplate` of a list gets the array of the page items (`[]` for an empty page; the whole response with `--raw`, `--withTotalPages` or `--withTotalElements`), a template that returns an array prints one document per element, csv, tsv and csvheader without `--select` use the `**` selection (empty objects and arrays are `{}` and `[]` columns, `null` prints as `null`), the `--debug` worker-loop lines and the order of `Creating c8y client`, the realtime and notification2 websocket log lines in the CLI log format (including `ws (recv)`), no `use of closed network connection` warning at the end of notification2 subscribe, the 10 second realtime handshake timeout, and `notification2 subscribe` exiting 0 on Ctrl-C (#838, go-c8y#247).
- Batch 4 restored these (go-c8y-cli #840 to #842, go-c8y #248 and #249), each measured against a `v2.59.0` build:
  - `--outputFile` and `--outputFileRaw` (#841): one compact JSON document per line with no final newline (csv and tsv write the rows, csvheader writes them without the header, `--outputTemplate` writes the template output), `{id}` expands per document, an empty list writes an empty file, no file for an error, `--dry` or `--output serverresponse` (`--withError` writes the error document), `--outputFileRaw` writes the whole response (last page) with `{id}`, `{name}`, `{type}`, `{owner}`, `{filename}` and `{basename}`, and an empty file for a response without a body;
  - `--outputTemplate` with `--dry` gets `request` (`host`, `method`, `path`, `pathEncoded`, `query`, `queryParams`, `url`; no `body`), a `response` with `statusCode` 0, an empty body and `header` `{}`, and `output` as an empty string; `response.header` is in the data of real requests again (#841);
  - page size (#841): a page size of blank, `0` or `5` (setting or `--pageSize`) sends no `pageSize` on list requests, also with `--totalPages N` and `--currentPage`; `--includeAll` still sends 2000;
  - lookup errors (#842, go-c8y#249): a transport error during a name lookup is a `commandError` (exit 101, per-resource text such as `Could not fetch by id: <err>`; exit 100 with the prefix `failed to lookup by name.` for a flag that does not drive the command), an unknown name on a secondary flag is `reference by name: no matches found`, unknown firmware, configuration, device group, smart group, device profile, agent and device service names report `referenceByName` (exit 4) instead of the SDK text, `software` and `firmware` versions and `firmware patches` given without the parent are looked up by version number, `ui applications plugins install|delete|update --plugin <id>` does a GET by id, `ui applications plugins --application <name>` looks the name up among hosted applications (owner first), and wildcard user names resolve to the first match on `users get|update|delete|resetUserPassword|revokeTOTPSecret` and `userreferences add|delete`; the name-not-found detection uses the typed SDK errors (`core.ErrNameNotFound`, `*core.LookupError`) instead of error text;
  - tooling (#840): the docs generator (`task docs-c8y`) no longer depends on the terminal (it sets `C8Y_FORCE_STDIN`, `C8Y_FORCE_STDOUT` and `C8Y_FORCE_STDERR_TTY`) and always documents the interactive defaults (`--output table`, `--view auto`, no default text for `-c` and `-M`); the `v2.59.0` generator was terminal dependent (453 pages differ with and without a tty). The stale pages `tenants disableApplication`, `enableApplication` and `listReferences` were removed: those commands are still hidden, as on `v2.59.0`.
- Not restored because `v2.59.0` was wrong (v2 bugs): `microservices serviceusers get --id <unknown name>` exits 0 with no output; `operations wait --id 1,2` reports only the last error; `--withError` prints the error document twice; an extension `binaryUploadURL` uploads for real under `--dry`; a transport error with `--includeAll` or `--totalPages` prints `commandError: Response is empty` (exit 101); `ui applications plugins list --application <unknown name>` exits 0 with no output (and exit 100 for an unknown numeric id); a piped absolute url in `c8y api` sends `pageSize` twice (`pageSize=20&pageSize=7`); `devices list --includeAll` sends one extra id keyset request after a short page; `--outputTemplate` with `--totalPages N` evaluates the template on the first page only and with `--includeAll` once for the first-page response; `realtime subscribe --verbose` logged the handshake token in clear text (v3 masks it as `{token}` while `logger.hideSensitive` is on, the default); piped input or `--workers` with `--outputFile` keeps only the last response (v3 writes every document); `--totalPages N` with `--outputFile` or `--outputFileRaw` uses the first page only (v3: all pages, and the last page for the raw file); `--outputTemplate` with `--dry` and `--includeAll` or `--totalPages N` prints no result (v3 prints one); `binaries get --outputFileRaw` also printed the file content to stdout (v3 does not); `--outputFile` with `{id}` and a list names one file after the first item (v3 writes one file per distinct path).

### Follows the v2 release (not a break)

v3 had not yet taken these v2.59.0 changes. They were merged into v3 (go-c8y-cli #827):

- `devicegroups create --type` accepts `c8y_DeviceSubgroup` (v2 #779). A script that used `c8y_DeviceSubGroup` is affected, as it is on v2.59.0.
- `c8y mcp serve --mode` is the ceiling, the highest mode the agent can reach (default `dev`). The user can raise the session mode up to it (v2 #812).
- `c8y mcp serve --sessionProvider` is new (v2 #781). The MCP server rejects unknown sub commands (#780), handles concurrent tool calls (#795) and classifies `applications open --noBrowser` as a read (#774).

### Restored behavior that v3 still differs from in one detail

- `--debug` prints the v2.59.0 lines again. The job iterator lines (`checking job iterator`, `finished adding jobs`) of a piped or `--id` run appear after the job on v3, because v3 runs serially, and v2.59.0 logged a second `Input line: %!s(<nil>)` for body commands (a format artifact): both are not restored. In `--debug` the output document appears after `job successful` on v3 and before it on v2.59.0 (stdout buffering).

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

- **Today**: realtime honors `--ignoreProxy`; notification2 and remote access always use the environment proxy. The CLI sets TLS, proxy and the handshake timeout for websockets in `pkg/cmdutil/wsconnector.go`, because the SDK builds its websocket dialer from the REST transport and the CLI wraps that transport, so the SDK cannot see through it. The realtime handshake timeout is 10 seconds as in `v2.59.0` (a connector field), and the cancel-aware dial `dialUntilDone` ends a connect when Ctrl-C is pressed, in realtime and notification2 subscribe (exit 2, `Interrupted`; `v2.59.0` waited 10 seconds in the bayeux handshake or was killed by the signal). A silent TCP server never made v3 wait 45 seconds, as an earlier version of this page claimed.
- **Proposed**: one rule for TLS, proxy and handshake timeout in all websocket commands. SDK side: an explicit websocket TLS and proxy option on `ClientOptions`, and an exported cancel-aware dial helper, so the CLI can drop the connector.
- **Breaks**: users who depend on the current proxy split (for example `--ignoreProxy` affecting only realtime).

### 5. One exit-code table and error format

- **Today**: 100 (unexpected error) and 101 (user error) are used inconsistently between built-in, extension and runtime commands. v3 now reproduces the `v2.59.0` codes and error text for built-in commands, including the inconsistencies below.
- **Proposed**: a single mapping (HTTP status codes, user errors, unresolved names, timeouts, assertion failures), documented and tested, with one error format. Known inconsistencies to clean up:
  - a denied `--device` file exits 101 for filter and body flags (and for the create commands of events, alarms, measurements and operations, which treat the flag as a file reference) and 100 for path arguments and assert commands; one rule for every flag that accepts a file is needed;
  - name lookup failures print `commandError: referenceByName: no matching items found ... ::StatusCode=404`, while a direct lookup (`usergroups getByName`) prints a `serverError` with the Error Response block; the SDK text (`managed object not found with name: x`) is clearer;
  - the remote access "not found" error has three texts and two exit codes: `configuration not found. name=` (101) for connect and server, `referenceByName: no matching items found. name=x  ::StatusCode=404` (4) for `configurations get`, `delete` and `update`; v3 matches `v2.59.0` (detected by the typed SDK error `*remoteaccess_configurations.ConfigurationNotFoundError`, go-c8y#249);
  - REST transport errors (connection refused, TLS verification failure) exit 99 with the raw Go message, printed twice, but 100 for the token request of `notification2 subscriptions subscribe` and for wait commands (`commandError`), and 112 for assert commands; v3 matches `v2.59.0`, including a transport error during a name lookup (`commandError`, exit 101, with a per-resource text: `Could not fetch by id`, `could not fetch microservices`, `could not fetch ui plugin`, `could not fetch applications`; exit 100 with `failed to lookup by name.` for a flag that does not drive the command). The misleading `Response is empty` of `v2.59.0` for `--includeAll` and `--totalPages` is not copied. In `v2.59.0` the 100 or 101 of a secondary flag depended on the generated command (a path parameter returned a plain error, exit 100; a query, header, form or body flag went through `NewUserError`, exit 101); v3 cannot tell and uses 100 for every secondary flag resolved while the call is built. Measured mismatches: `configuration send --configuration <unknown>` (v2 101, v3 100), `firmware|software versions install` with an unknown `--firmware`, `--software` or `--device` (v2 101 `item: not found`, v3 100 `reference by name: no matches found`), a refused `firmware versions create --firmware <name>` (v2 100, v3 101) and `inventory assets assign --id <unknown name>` (v2 did no lookup, exit 4; v3 100). A full restore needs the generator to mark path flags. The `--withError` JSON of the `notification2` token error has `"exitCode":100` on v3, which `v2.59.0` omits. One rule (for example "could not reach the server", one exit code, the message once) would be clearer;
  - an error whose message equals the server message repeats the text after `METHOD url: status type`;
  - `--withError` prints the error document twice on `v2.59.0` (sorted and nested);
  - the aggregate "jobs completed with N errors" line depends on the worker type, not on the command kind;
  - `operations wait --id 1,2` on `v2.59.0` reports only the last error and exits 100;
  - `microservices serviceusers get --id <unknown name>` exits 0 with no output on `v2.59.0`;
  - error text for a response without a JSON body: v1 printed an empty message, v3 derives it from the SDK;
  - the v1 `details` field of the error JSON (debug mode only) is not restored.
  - a missing required parameter of an extension command: exit 100 for a path parameter, 101 for a query or body parameter, with two texts (`missing required parameters. [device]` for reference types, `Missing required parameter. id` for the others). An extension body parameter that is pipeline and required prints `Body is missing required properties: source.id` on `v2.59.0` and `Missing required parameter. device` on v3 (both exit 101);
  - name-not-found has three texts: `reference by name: no matches found` (flag; exit 100 for a path flag, 101 for a body or query flag in `v2.59.0`), `referenceByName: no matching items found. name=x  ::StatusCode=404` (pipeline or path, exit 4) and the SDK text `application not found with name: x`; one text and one exit code, with the candidate list, would be clearer. Firmware, configuration, device group, smart group, device profile, agent and device service names now report `referenceByName` like the others (before #842 the SDK text was printed);
  - one rule for a failed lookup ("could not reach the server", one text, one exit code) instead of the per-resource texts and the 100/101 split of path and body flags;
  - `ui applications plugins list --application <unknown>` exits 0 with no output on `v2.59.0`; an error would be clearer (v3 already exits 4 for an unknown name);
- **Breaks**: scripts that match the exit codes or messages that change.
- **Source**: CLI#828, CLI#831.

### 6. Name resolution: wildcards and ambiguous matches

- **Today (v3, as `v2.59.0`)**: extension `user[]` and `userself[]` expand wildcards and take the first match (`--user 'adm*'`; the server filter is a name prefix with the wildcards removed, over a page of 5 users, so `a*b` filters by `ab`; no match gives `reference by name: no matches found`). `hostedapplication` matches hosted applications only (current tenant first, exact name with no wildcards, a numeric id is used as is). The SDK's `Users.ResolveID` supports wildcards but returns an "ambiguous" error when more than one user matches. Role lookups keep wildcard and first-match.
- **Also (name lookups, #842)**: device name lookups do not filter on `has(c8y_IsDevice)` (`v2.59.0` did), and use `pageSize=1` with `$orderby=name,creationTime` (`v2.59.0`: `pageSize=5`, server order). For an ambiguous name v3 picks the oldest by creation time, `v2.59.0` took the first of an undefined server order. A non-device object with the same name as a device can be picked. Restoring needs a device-scoped SDK reference (about 180 CLI call sites pass `NameOrID`). `software|firmware versions` and `firmware patches` given without `--software` or `--firmware` pick the first version with that number of any software or firmware (as `v2.59.0`); requiring the parent, or failing on ambiguity, would be safer.
- **Proposed**: one rule for all name lookups: wildcards allowed; more than one match is an error that lists the candidates (safer for update and delete), with an opt-in `--first` for the old behavior; device names match devices only.
- **Breaks**: scripts that rely on the first match.

### 7. Login TLS verification

- **Today (kept on purpose)**: BASIC, OAUTH2_INTERNAL and OAUTH2 login and the login-options lookup accept self-signed certificates whatever `--allowInsecure` says, as the v1 login client did. The SDK flows (BROWSER, DEVICE, CERTIFICATE) honor `--allowInsecure`.
- **Proposed**: verify certificates unless `--allowInsecure` is set.
- **Breaks**: users with self-signed tenants who never set `--allowInsecure`.
- **Ways to ship**: warn when a login succeeds only because verification was skipped, then flip the default.

### 8. Default page size and the stray `pageSize`

- **Today**: the CLI sends `pageSize=20` unless the setting is blank (a space) or `0`; an empty string still sends 20. An explicit `5` (setting or `--pageSize 5`) sends no `pageSize` at all, because of a quirk inherited from the v1 SDK (the SDK default is 5). Get-by-id requests (`devices get`, `inventory get`, count commands and similar GET requests) carry `?pageSize=20` on `v2.59.0`; v3 sends it again through a request hook on the call context, so the `--verbose` line and the error message match. Commands that send none on `v2.59.0` are exempt (`NoDefaultPageSize`): wait and assert commands, extension runtime commands, `microservices serviceusers get`, `ui applications plugins list|install|delete|update|replace`, `ui plugins create`, `deviceregistration register-basic|register-ca`, `devices enroll`, `remoteaccess connect ssh|run` and `server`, and the v3-only datahub commands (`datahub jobs get` sends `pageSize` and `datahub jobs listresults` sends it as `limit`, as in `v2.59.0`). With a blank page size, `0` or `5` the list commands send none, as `v2.59.0`, also with `--totalPages N` and `--currentPage` (N pages of the platform default, 5 items); the CLI strips the platform default with a request hook, which never changes the result. A list with `--includeAll` sends 2000. With no settings file at all an unset page size behaves like blank (no `pageSize`); the 20 comes from the settings file written by the installer. `devicemanagement certificate-authority get` sends `pageSize=20` on v3 (`v2.59.0`: `pageSize=5&withTotalElements=true`).
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

### 15. What `--raw` prints

- **Today**: `--raw` prints the response as received for an enveloped list, and a top-level array as an array (`[]` when empty), but `--select` and `--flatten` split a top-level array into one document per item, and `--filter` keeps the array (`v2.59.0`, restored on v3 by CLI#826 and CLI#829). A filter on an enveloped list under `--raw` prints nothing on both. Every command whose response is a plain array or an envelope prints it as received with `--raw` (13 commands returned items before #838), and `ui applications plugins list` prints one document per plugin, also with `--raw`.
- **Proposed**: one rule: `--raw` always prints the unmodified response (envelope or bare array), and `--select`, `--flatten` and `--filter` either do not apply under `--raw` or always apply to the items.
- **Breaks**: scripts that rely on the `v2.59.0` mix of shaped items and arrays under `--raw`.
- **Ways to ship**: opt-in setting first, then flip the default in a major release.
- **Source**: CLI#820, CLI#826, CLI#829.

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

- **Today**: `--verbose` and `--debug` print the `v2.59.0` line format (no HTTP dump, no SDK log lines). Output replaces the escapes `\u003c`, `\u003e` and `\u0026` with the characters `<`, `>` and `&` in all output formats, with a byte-level replace over the whole text (an escaped backslash followed by `u003c` is also altered on `v2.59.0`; v3 replaces only the real escapes in that case so the JSON stays valid). Under `--workers N`, the worker number in `--verbose` is the SDK pool slot, not the v2 worker id.
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

- **Today**: `--totalPages N` on `inventory list` and on the time-series lists (events, alarms, measurements, operations) reads N pages by page number, as `v2.59.0` (the first request has no `currentPage`); the time keyset is used only for `--includeAll` or an explicit `--paginationStrategy time`. The typed inventory lists (devices, agents, devicegroups, smartgroups, deviceprofiles, configuration, software, firmware and their versions and patches) still differ for N above 1: `v2.59.0` used the id keyset (first request `_id gt '0'`, `$orderby=_id asc`, results ordered by id), v3 keeps `$orderby=name` (or `--orderBy`) and offset paging, and N pages still come back. The SDK id keyset exists only for devices, agents, devicegroups, smartgroups, deviceprofiles and configuration lists, so the firmware and software lists need an SDK change. `inventory find` uses the id keyset in both, but the later-page filter differs (`v2.59.0` nests `_id gt 'N' and (_id gt '0' and (...))`, v3 sends `_id gt 'N' and (...)`); the results are the same. `--paginationStrategy` (auto, offset, id) exists only on v3. In the SDK the offset, id keyset and time keyset strategies no longer send `currentPage` on the first request (id keyset: on any request), and the sequential walk no longer forces `withTotalPages` or `withTotalElements`; this matches `v2.59.0`, but SDK callers that relied on it see a change.
- **Decision (2026-11-28)**: v3 keeps the name order for `--totalPages N` on the typed inventory lists. The id keyset of `v2.59.0` is not restored. This is a breaking change to list in the v3 release notes.
- **Proposed**: decide whether the SDK strategy is the rule everywhere, and document the order and the keyset.
- **Breaks**: scripts that depend on the order of results with `--includeAll`, and scripts that expect results ordered by id with `--totalPages N` on the typed inventory lists (the name order is kept on v3).

### 21. Object type of `binaries create` for a file without extension

- **Today**: v3 matches `v2.59.0`: the object type is the type of the source file's extension, else `application/octet-stream`, with no content sniffing, and `--name` does not change it. In the SDK, an upload from a reader is typed by the extension of its name. One difference remains: `--type my/type` also types the file part on v3 (SDK `UploadFileOptions.ContentType`), while `v2.59.0` set only the object type.
- **Proposed**: sniff the content (what the SDK designed) or keep octet-stream; and decide whether `--type` types the file part.
- **Breaks**: scripts that match the type of binaries created from extensionless files, or the part type of an upload with `--type`.

### 22. Per-request timeout model

- **Today**: `--timeout` is restored. A real deadline that starts before the request is prepared (the effect of `--timeout 1ms` on `v2.59.0`, a race even there: 0 to 3 of 10 runs timed out on loopback) is not reproduced, so a timeout test needs a slow server.
- **Proposed**: define when the timeout starts (request preparation, connect, whole command).
- **Breaks**: scripts that depend on very short timeouts.

### 23. Selected output details that are per document or per page

- **Today**: csv, tsv and csvheader columns are per document on `v2.59.0` and v3: a document with extra fragments adds columns to its row. `--outputTemplate` with `--totalPages N` evaluates the template on the first page only on `v2.59.0` and on every page on v3; with `--includeAll` once for the first-page response on `v2.59.0` and per page on v3 (v3 does not copy the `v2.59.0` behavior, which looks like a bug). `--raw` plus `--select` or `--filter` on a bare array still mixes shaped items and arrays (topic 15). Ctrl-C during the connect of a realtime or notification2 subscribe exits 2 with `Interrupted`. `inventory parents get` lacks the collection property annotation (only the `--debug` and `--verbose` array logs differ).
- **Proposed**: a stable column set for csv (a first-document or view-based set), one rule for templates with paging, one exit code for an interrupted connect.
- **Breaks**: scripts that parse csv columns or depend on the current template evaluation.
- **Also**: `v2.59.0` trimmed the trailing tabs of the last tsv row; v3 keeps every row consistent (a view parity detail from #835).
- **View follow-ups (go-c8y-cli#835, #840)**:
  - table output with `--flatten` shows the nested values on v3; `v2.59.0` showed empty cells (a v3 fix, to confirm);
  - column union: collect the columns over several documents (the first page, or all documents of the stream) so that csv, csvheader and table have one column set. Today the view of the first document decides, and when it matches no view the csv rows can have different cell counts (the same on `v2.59.0`). Proposal: an opt-in setting that unions the columns of the buffered documents; it needs buffering, so it does not suit a long stream;
  - an unknown `--view` name warns (`no matching view found. name=nope`), shows all columns and exits 0; it could exit non-zero;
  - json with a view keeps only the properties a document has; it could print empty values for missing columns as table and csv do;
  - detection could repeat per response or page until a view matches (`v2.59.0` detects once).
- **Output files (#841)**: v3 keeps behavior where `v2.59.0` was wrong (see the note on restored behavior): every document with piped input or `--workers`, all pages with `--totalPages N` (the last page for `--outputFileRaw`), one file per distinct `{id}` or `{name}` path for a list. Restored as `v2.59.0` but arguably worth changing: `csvheader` writes no header row to the file, and `--output serverresponse` writes no file. The documented behavior of `--outputFile` should be decided (one JSON document per line is the `v2.59.0` format).

### 24. Request shape of count commands and `c8y api` urls

- **Today**: assert count commands request `withTotalPages` as `v2.59.0` does (v3 sent `withTotalElements`, and `revert=false` for measurements, before #837). `c8y api` with a piped absolute url sends a single `pageSize` on v3 (a duplicate on `v2.59.0`).
- **Proposed**: use `withTotalElements` for the count (cheaper, clearer).
- **Breaks**: nothing visible, only the request shape.

## Part 2: differences that remain on v3 (confirm or revert)

These differences from `v2.59.0` exist on v3 today (measured on the fake server unless a source is named). A v2 CLI user would notice them. Each should be confirmed (and listed in the release notes) or reverted.

### Exit codes and errors

| Change | v2.59.0 | v3 | Source |
|---|---|---|---|
| `microservices serviceusers get --id <unknown name>` | exit 0, no output | exit 4, `referenceByName: no matching items found` | CLI#799 |
| `operations wait --id 1,2` with several failing ids | exit 100, only the last error | exit 104, one error per id and `jobs completed with 2 errors` | CLI#828 |
| `--withError` with a failing request | prints the error document twice | prints it once | CLI#828 |
| `ui applications plugins list --application <unknown name>` | exit 0, no output (exit 100 for an unknown numeric id) | exit 4 `referenceByName` (the name is looked up among hosted applications, owner first, as `v2.59.0`); a server error for an unknown numeric id | CLI#836, CLI#842 |

### Behavior

- **Dry run is more complete**: `ui plugins create --file <url>` downloads the url under `--dry` (v2 failed with "no such file"). An extension `binaryUploadURL` parameter no longer uploads for real under `--dry` (v2 did) and shows a placeholder url in the body (CLI#805).
- **`--withTotalPages` and `--withTotalElements` with `--select`**: v2 returned the whole response, with the statistics, and `--select` paths start at its top level. v3 prints each item on its own line, so the statistics are lost and `--select statistics.totalElements` matches nothing. This is deferred on purpose: the go-c8y v2 client has to keep the statistics with the results first, and the CLI has no workaround. `c8y api` is not affected.
- **Multi-value flags**: `microservices serviceusers get --id` accepts a list (CLI#799).
- **`c8y api` with a piped absolute url that has its own `pageSize`**: `v2.59.0` sends the parameter twice (`pageSize=20&pageSize=7`), v3 sends a single `pageSize=20` (a v2 bug, not copied).
- **`--totalPages N` on the typed inventory lists** (devices, agents, devicegroups, smartgroups, deviceprofiles, configuration, software, firmware and their versions and patches) keeps `$orderby=name` and offset paging on v3 (`v2.59.0` used the id keyset ordered by `_id`); N pages still come back. Kept on purpose (decision of 2026-11-28, topic 20); list it in the release notes.
- **Not restored, logging**: after the end of a realtime stream (duration, count, Ctrl-C) `v2.59.0` logged nothing from the disconnect, and v3 mutes the websocket log from that point (two clean-up lines of the SDK are dropped for notification2). The order of `Pending messages ids` and of `Received message` in `realtime subscribeAll` is random on both versions.
- **Output files**: v3 writes every document of piped input or `--workers`, all pages of `--totalPages N`, and one file per distinct `{id}` path (v2 bugs, see the restored note). A template error after a successful response with `--outputFile` leaves an empty file. `--outputFile` without `--select` for a get writes compact JSON; `v2.59.0` wrote the body bytes as received (the same for real Cumulocity responses). The order of the `--workers` file result is not reproduced. `--outputTemplate` with `--dry` and `--includeAll` or `--totalPages N` prints one result (`v2.59.0`: none). `--withTotalPages` and `--withTotalElements` with `--outputFile` follow stdout (items only).
- **Other differences found while restoring requests**: name lookups use `pageSize=1` with `$filter=... $orderby=name,creationTime` instead of `pageSize=5` (topic 6); device name lookups lack `has(c8y_IsDevice)`; `applications createHostedApplication` looks the application up with a name filter and `pageSize=20` (`v2.59.0`: `pageSize=2000`, client-side exact match; same result); `ui plugins` name lookups request `hasVersions=true&type=HOSTED` without `owner=t1` first (only differs for same-named plugins in several tenants); `devices services get|update|delete` look the service up under `/inventory/managedObjects/<device>/childAdditions` with `pageSize=1` (`v2.59.0`: a `managedObjects` query with `pageSize=5`); `software versions install` lists the software twice by name (same output); `inventory get --id <name>` resolves the name (`v2.59.0`: a GET by id, 404); `ui applications plugins install|delete --application <id>` with an unreachable server prints the plain `serverError` (`v2.59.0`: prefix `failed to find hosted application.`, exit 99 in both); `software versions get --id <version>` with an unreachable server prints `Could not fetch by id` (`v2.59.0`: `Fetcher could not build a valid query`, exit 101 in both). `currenttenant listApplications --raw` prints nothing against the fake server, whose tenant has no `applications` property (probably a fake-server artifact).

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
- Typed lookup errors (#249): `core.ErrNameNotFound`, `*core.NameNotFoundError` (`Resource`, `Name`; still a 404), `core.NameNotFound(err)`, `*core.LookupError` (wraps every error while a reference is resolved; `api.IsLookupError`) and `remoteaccess_configurations.ErrConfigurationNotFound` / `*ConfigurationNotFoundError` (matches `core.ErrNameNotFound` but is not a 404, following `ErrPassthroughNotFound`; making it a 404 would be more consistent). `ResolveID` of agents, device and smart groups, device profiles, configurations, user roles and trusted certificates now returns a 404 `NameNotFoundError` for an unknown name (before: a plain error). The name lookups of applications, microservices, UI plugins and the plugin catalog no longer hide a failed list request as "not found with name". `UIApplicationPlugins.Update` with an id does one GET by id. Messages are unchanged.
- `RetryOptions.Exponential` and additive identity headers (a custom `User-Agent` or `X-Application` is kept) are new (#245). `PaginationOptions.SinglePage` sends the plain request (#243); its doc says the request is "unchanged", but `PageSize` still goes through `OptimalPageSize` (with `PageSize` 0 it becomes `MaxItems` or 2000, and the CLI relies on this to send `pageSize=20`).
- `api.WithNotFoundAsError` and `api.WithConflictAsError` only change the SDK's own `IgnoreNotFound`/`IgnoreConflict` calls. A caller's explicit `.IgnoreStatusCodes(404)` still clears the error, which is intended (the caller asked for it) but worth documenting next to the options.
- `Microservices.Delete` with a name that does not resolve returns Skipped ("not found") even with `api.WithNotFoundAsError`, because the not-found comes from the name lookup, not from a 404 on the delete. The CLI resolves the name itself to keep exit 4 (CLI#811). Making the SDK honor the option here is a fix for callers who set it, and a change for callers who rely on Skipped. With the typed `core.ErrNameNotFound` this no longer needs a text match.
- `Registration.Register` on an existing external id used to fail against the fake server ("already exists"). With a new password it now succeeds and updates the credentials, like the platform (the response lists the id in `credentialUpdatedList`); with the same password the row still fails ("Error during password validation."). Callers that relied on the error must check `credentialUpdatedList` instead. The CLI test `tests/manual/deviceregistration/register_bulk.yaml` covers this.
- The fake server's 404 error type changed from `<resource>/notFound` to the platform style `<resource>/Not Found` for managed objects, alarms, events, measurements and operations. Code that matches the old type string breaks.

## Part 3: still to verify

- `sessions login --tfaCode` is now used (it was ignored on the old login path), and `sessions set --tfaCode` is a code, not a TOTP secret (CLI#804). Not run against `v2.59.0` here, because it needs a tenant with two-factor authentication.
- The first command of a session no longer makes an extra request, because the tenant is taken from the token when possible (CLI#804). Not counted against `v2.59.0` here.
- The aggregate "jobs completed with N errors" line was compared for assert commands only. Other generic-worker commands (certificate authority, register-bulk, template execute, sessions encrypt and decrypt) print no line on `v2.59.0` and were not checked on v3.
- The fileguard rule (exit 101 or 100) depends on the flag name and command path, and was compared on about 30 commands. Other commands with a `device` driver flag could differ.
- Name lookup detection uses the typed SDK errors (`core.ErrNameNotFound`, `*core.LookupError`); only the `groupByName` 404 check of the url path remains, and unit tests cover that a 404 matched only by its text is not treated as a name lookup. The "flag does not drive the command" decision (exit 100 versus 101) is a heuristic (the flag read last before the failing lookup, compared with the driving flag's value). It is exact for the measured commands (remote access configurations, software and firmware versions and patches, userroles and userreferences groups, devices services), not for every command that resolves a second reference inside the SDK call.
- The TLS cases of the transport errors (`tests/offline/common/transport_tls.sh`) print a pass line without TLS. They run in CI through `test-cli-offline-manual`, which runs `errors_parity.yaml` with `C8Y_OFFLINE_TLS=1`; the TLS cases in `tests/offline/realtime/subscribe_parity.yaml` also only mean something under `C8Y_OFFLINE_TLS=1`, which `task test-cli-offline-realtime` sets.
- The retry timing test in `tests/offline/common/output_parity.yaml` checks wall-clock waits and could be flaky on a heavily loaded runner.
- A custom `--header "User-Agent: x"` or `--header "X-Application: x"` on the realtime and notification2 websocket handshake (the SDK sets both headers itself there) was not compared with `v2.59.0`.
- Confirmation prompts of internal requests: a command with a read semantic method (`flags.WithSemanticMethod("GET")`) no longer asks, also with `--confirm`, before the POST/PUT/DELETE it sends itself (`notification2 subscriptions subscribe` sent a token request that asked, `devices enroll` and `deviceregistration getCredentials` already had the annotation). Measured without `C8Y_SETTINGS_CI` and with `C8Y_MODE=dev`: the realtime, alarms, events, operations and other subscribe commands, `remoteaccess`, `sessions login` (own client, no approver), name lookups (the approver is removed while a name is resolved) and `api GET` send no data-changing request through the approver; `devices delete`, `inventory create`, `alarms create`, `api POST`, `currentuser logout`, `notification2 tokens create` and `subscriptions create` still ask, as on `v2.59.0`.
- Two differences from `v2.59.0` were kept: `microservices serviceusers create` created the application without asking on `v2.59.0` (a create command that skipped the prompt) and asks on v3; `--confirm` on a plain read (`devices list --confirm`) asked on `v2.59.0` and does not ask on v3, because the SDK approver only sees requests that change data.
- go-c8y#250 adds `realtime.ConnectError{URL, Err}` (message unchanged), and the CLI recognizes a failed realtime websocket with `errors.As`. The notification2 client returns the dial error without a prefix, so there is no typed error for it.
- The CLI `go.mod` is pinned to go-c8y `199203f821ec` (the merge of #249; #248, `1fac089f`, only changes a test). Re-pin to the newest `origin/v2` commit when another SDK PR merges.
- go-c8y#249 was merged while the external SonarCloud quality gate failed on duplicated new lines (8.8%, limit 3%; the new lines sit in blocks that are already copy-pasted across the SDK services, as in #243). All GitHub Actions jobs passed.
- Output and lookup parity was measured on a fake server and a request-logging stub, not on a real tenant. The TLS variant of `lookup_errors_parity.yaml` was not run, and the `--workers` ordering of `--outputFile` was not compared. The parity row "unset pageSize sends `pageSize=20`" depends on the settings file (topic 8).
- The `NoDefaultPageSize` exemption list (topic 8) comes from the commands that were compared. Other hand-written commands that send a GET without a page size on `v2.59.0` could still show a spurious `pageSize`; a differential run of the 283 single-line `tests/auto` commands found none. `devices enroll` and other manual commands were not compared beyond the listed ones.
- `go-c8y#247` merged with a failing `test-live` job (`Test_Registration_Register`, "Error during password validation"). go-c8y#248 gives the test a policy-compliant password prefix (`Pass-7x!`), as the bulk test already had. The cause was inferred from the code and logs, not proven (the failure was intermittent); `test-live` passed on that PR.

## Notes and open questions (not CLI behavior)

- **Live tests that need a tenant feature**: a live test cannot skip itself when the tenant lacks a feature. For example, Notification 2.0 answers the test user with 403 (`security/Forbidden`), so the test fails although the CLI works. This is a convention for the test suite, not a CLI change. Proposed: a shared helper in `tests/manual` that probes the feature once (a cheap request, a 403 or 404 meaning "not available") and marks the test as skipped, and a list of the features that the shared test tenant has. Decision needed: where the helper lives (a script sourced by the tests, or a commander precondition) and whether a skipped test counts as a pass in the live workflow summary. Source: CLI#824.
- **Offline test suites share a temp folder**: the cache suite uses `$TMPDIR/go-c8y-cli-cache`, so two runs at the same time clash. Use a private `TMPDIR` per run.
- **Generated command docs**: regenerated, and the generator is terminal independent (#840). Run `task docs-c8y` after a spec change; it removes the old pages first.
- **Hand-written commands skip the generator**: `devicegroups create` on v3 is migrated by hand, so a spec change must also be made in `pkg/cmd/*/*_create.go`.
- **Merge v2 into v3 with a merge commit**, not a squash, so `origin/v2` stays an ancestor and later catch-ups stay small.
