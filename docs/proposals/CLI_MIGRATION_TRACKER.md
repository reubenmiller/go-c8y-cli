# CLI → v2 migration tracker (per-command)

One row per individual go-c8y-cli command still on a spec-generated `*.auto.go`, cross-referenced against the go-c8y **v2** SDK. Fill the **Pri** (priority) column to plan; sorted by effort then SDK-readiness but meant to be re-sorted/filtered freely. Regenerate with `python3 docs/proposals/gen_tracker.py`.

**Remaining commands: 18.**  Effort — S:5 · M:12 · L:1.  SDK call — ✅ ready:5 · ⚠️ service-exists-method-missing:13 · ❌ no-service:0.

**Effort:** `S` ≈ ≤½ day (SDK method ready, mechanical) · `M` ≈ ~1 day (new SDK method / resolution / query-build / sub-resource) · `L` ≈ multi-day (new SDK service, or binary/multipart).

**SDK call:** ✅ `Service.Method` exists · ⚠️ service exists but this method must be added (`Service?` = surface unverified) · ❌ no service at all.

| Pri | Command | HTTP | Endpoint | Effort | SDK call |
|:--:|---|:--:|---|:--:|---|
|  | `c8y microservices delete` | DELETE | `/application/applications/{id}` | S | ✅ Microservices.Delete |
|  | `c8y microservices get` | GET | `/application/applications/{id}` | S | ✅ Microservices.Get |
|  | `c8y microservices list` | GET | `/application/applications` | S | ✅ Microservices.List |
|  | `c8y microservices update` | PUT | `/application/applications/{id}` | S | ✅ Microservices.Update |
|  | `c8y notification2 subscriptions` | — | `— (dynamic)` | S | ✅ Notification2.List |
|  | `c8y microservices disable` | DELETE | `/tenant/tenants/{tenant}/applications/{id}` | M | ⚠️ Microservices +method |
|  | `c8y microservices enable` | POST | `/tenant/tenants/{tenant}/applications` | M | ⚠️ Microservices +method |
|  | `c8y microservices getbootstrapuser` | GET | `/application/applications/{id}/bootstrapUser` | M | ⚠️ Microservices +method |
|  | `c8y microservices getstatus` | GET | `/inventory/managedObjects?type=c8y_Application_{id}` | M | ⚠️ Microservices +method |
|  | `c8y microservices loglevels` | — | `— (dynamic)` | M | ⚠️ Microservices +method |
|  | `c8y notification2 tokens` | — | `— (dynamic)` | M | ⚠️ Notification2 +method |
|  | `c8y tenants applications` | — | `— (dynamic)` | M | ⚠️ Tenants +method |
|  | `c8y tenants disable` | PUT | `/tenant/tenants/{id}` | M | ⚠️ Tenants +method |
|  | `c8y tenants disableapplication` | DELETE | `/tenant/tenants/{tenant}/applications/{application}` | M | ⚠️ Tenants +method |
|  | `c8y tenants enable` | PUT | `/tenant/tenants/{id}` | M | ⚠️ Tenants +method |
|  | `c8y tenants enableapplication` | POST | `/tenant/tenants/{tenant}/applications` | M | ⚠️ Tenants +method |
|  | `c8y ui plugins` | — | `— (dynamic)` | M | ⚠️ UIPlugins? |
|  | `c8y microservices createbinary` | POST | `/application/applications/{id}/binaries` | L | ⚠️ Microservices +method |
