# CLI → v2 migration tracker (per-command)

One row per individual go-c8y-cli command still on a spec-generated `*.auto.go`, cross-referenced against the go-c8y **v2** SDK. Fill the **Pri** (priority) column to plan; sorted by effort then SDK-readiness but meant to be re-sorted/filtered freely. Regenerate with `python3 docs/proposals/gen_tracker.py`.

**Remaining commands: 8.**  Effort — S:1 · M:7 · L:0.  SDK call — ✅ ready:1 · ⚠️ service-exists-method-missing:7 · ❌ no-service:0.

**Effort:** `S` ≈ ≤½ day (SDK method ready, mechanical) · `M` ≈ ~1 day (new SDK method / resolution / query-build / sub-resource) · `L` ≈ multi-day (new SDK service, or binary/multipart).

**SDK call:** ✅ `Service.Method` exists · ⚠️ service exists but this method must be added (`Service?` = surface unverified) · ❌ no service at all.

| Pri | Command | HTTP | Endpoint | Effort | SDK call |
|:--:|---|:--:|---|:--:|---|
|  | `c8y notification2 subscriptions` | — | `— (dynamic)` | S | ✅ Notification2.List |
|  | `c8y notification2 tokens` | — | `— (dynamic)` | M | ⚠️ Notification2 +method |
|  | `c8y tenants applications` | — | `— (dynamic)` | M | ⚠️ Tenants +method |
|  | `c8y tenants disable` | PUT | `/tenant/tenants/{id}` | M | ⚠️ Tenants +method |
|  | `c8y tenants disableapplication` | DELETE | `/tenant/tenants/{tenant}/applications/{application}` | M | ⚠️ Tenants +method |
|  | `c8y tenants enable` | PUT | `/tenant/tenants/{id}` | M | ⚠️ Tenants +method |
|  | `c8y tenants enableapplication` | POST | `/tenant/tenants/{tenant}/applications` | M | ⚠️ Tenants +method |
|  | `c8y ui plugins` | — | `— (dynamic)` | M | ⚠️ UIPlugins? |
