# CLI → v2 migration tracker (per-command)

One row per individual go-c8y-cli command still on a spec-generated `*.auto.go`, cross-referenced against the go-c8y **v2** SDK. Fill the **Pri** (priority) column to plan; sorted by effort then SDK-readiness but meant to be re-sorted/filtered freely. Regenerate with `python3 docs/proposals/gen_tracker.py`.

**Remaining commands: 83.**  Effort — S:32 · M:40 · L:11.  SDK call — ✅ ready:32 · ⚠️ service-exists-method-missing:42 · ❌ no-service:9.

**Effort:** `S` ≈ ≤½ day (SDK method ready, mechanical) · `M` ≈ ~1 day (new SDK method / resolution / query-build / sub-resource) · `L` ≈ multi-day (new SDK service, or binary/multipart).

**SDK call:** ✅ `Service.Method` exists · ⚠️ service exists but this method must be added (`Service?` = surface unverified) · ❌ no service at all.

| Pri | Command | HTTP | Endpoint | Effort | SDK call |
|:--:|---|:--:|---|:--:|---|
|  | `c8y devices getchild` | GET | `inventory/managedObjects/{device}/childDevices/{reference}` | S | ✅ Devices.Get |
|  | `c8y devices getsupportedmeasurements` | GET | `inventory/managedObjects/{device}/supportedMeasurements` | S | ✅ Devices.ListSupportedMeasurements |
|  | `c8y devices getsupportedseries` | GET | `inventory/managedObjects/{device}/supportedSeries` | S | ✅ Devices.ListSupportedSeries |
|  | `c8y devices listassets` | GET | `inventory/managedObjects/{id}/childAssets` | S | ✅ Devices.List |
|  | `c8y devices listchildren` | GET | `inventory/managedObjects/{device}/childDevices` | S | ✅ Devices.List |
|  | `c8y inventory create` | POST | `inventory/managedObjects` | S | ✅ ManagedObjects.Create |
|  | `c8y inventory delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Delete |
|  | `c8y inventory findbytext` | GET | `inventory/managedObjects` | S | ✅ ManagedObjects.List |
|  | `c8y inventory get` | GET | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Get |
|  | `c8y inventory list` | GET | `inventory/managedObjects` | S | ✅ ManagedObjects.List |
|  | `c8y inventory update` | PUT | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Update |
|  | `c8y microservices delete` | DELETE | `/application/applications/{id}` | S | ✅ Microservices.Delete |
|  | `c8y microservices get` | GET | `/application/applications/{id}` | S | ✅ Microservices.Get |
|  | `c8y microservices list` | GET | `/application/applications` | S | ✅ Microservices.List |
|  | `c8y microservices update` | PUT | `/application/applications/{id}` | S | ✅ Microservices.Update |
|  | `c8y notification2 subscriptions` | — | `— (dynamic)` | S | ✅ Notification2.List |
|  | `c8y software create` | POST | `inventory/managedObjects` | S | ✅ Repository.Software.Create |
|  | `c8y software delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ Repository.Software.Delete |
|  | `c8y software get` | GET | `inventory/managedObjects/{id}` | S | ✅ Repository.Software.Get |
|  | `c8y software list` | GET | `inventory/managedObjects` | S | ✅ Repository.Software.List |
|  | `c8y software update` | PUT | `inventory/managedObjects/{id}` | S | ✅ Repository.Software.Update |
|  | `c8y users create` | POST | `user/{tenant}/users` | S | ✅ Users.Create |
|  | `c8y users delete` | DELETE | `user/{tenant}/users/{id}` | S | ✅ Users.Delete |
|  | `c8y users get` | GET | `user/{tenant}/users/{id}` | S | ✅ Users.Get |
|  | `c8y users getinventoryrole` | GET | `/user/inventoryroles/{id}` | S | ✅ Users.Get |
|  | `c8y users getuserbyname` | GET | `user/{tenant}/userByName/{name}` | S | ✅ Users.GetByUsername |
|  | `c8y users list` | GET | `/user/{tenant}/users` | S | ✅ Users.List |
|  | `c8y users listinventoryroles` | GET | `/user/inventoryroles` | S | ✅ Users.List |
|  | `c8y users listusermembership` | GET | `/user/{tenant}/users/{id}/groups` | S | ✅ Users.ListGroupsWithUser |
|  | `c8y users resetuserpassword` | PUT | `user/{tenant}/users/{id}` | S | ✅ Users.ResetPassword |
|  | `c8y users revoketotpsecret` | DELETE | `user/{tenant}/users/{id}/totpSecret/revoke` | S | ✅ Users.Update |
|  | `c8y users update` | PUT | `user/{tenant}/users/{id}` | S | ✅ Users.Update |
|  | `c8y bulkoperations listoperations` | GET | `devicecontrol/operations` | M | ⚠️ Bulkoperations +method |
|  | `c8y currentapplication get` | GET | `/application/currentApplication` | M | ❌ none |
|  | `c8y currentapplication listsubscriptions` | GET | `/application/currentApplication/subscriptions` | M | ❌ none |
|  | `c8y currentapplication update` | PUT | `/application/currentApplication` | M | ❌ none |
|  | `c8y devicemanagement certificates` | — | `— (dynamic)` | M | ⚠️ TrustedCertificates +method |
|  | `c8y devices assignchild` | POST | `inventory/managedObjects/{device}/childDevices` | M | ⚠️ Devices +method |
|  | `c8y devices availability` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices children` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices services` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices statistics` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices unassignchild` | DELETE | `inventory/managedObjects/{device}/childDevices/{childDevice}` | M | ⚠️ Devices +method |
|  | `c8y devices user` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y inventory additions` | — | `— (dynamic)` | M | ⚠️ ManagedObjects +method |
|  | `c8y inventory assets` | — | `— (dynamic)` | M | ⚠️ ManagedObjects +method |
|  | `c8y inventory children` | — | `— (dynamic)` | M | ⚠️ ManagedObjects +method |
|  | `c8y inventory count` | GET | `inventory/managedObjects/count` | M | ⚠️ ManagedObjects +method |
|  | `c8y inventory find` | GET | `inventory/managedObjects` | M | ⚠️ ManagedObjects +method |
|  | `c8y microservices disable` | DELETE | `/tenant/tenants/{tenant}/applications/{id}` | M | ⚠️ Microservices +method |
|  | `c8y microservices enable` | POST | `/tenant/tenants/{tenant}/applications` | M | ⚠️ Microservices +method |
|  | `c8y microservices getbootstrapuser` | GET | `/application/applications/{id}/bootstrapUser` | M | ⚠️ Microservices +method |
|  | `c8y microservices getstatus` | GET | `/inventory/managedObjects?type=c8y_Application_{id}` | M | ⚠️ Microservices +method |
|  | `c8y microservices loglevels` | — | `— (dynamic)` | M | ⚠️ Microservices +method |
|  | `c8y notification2 tokens` | — | `— (dynamic)` | M | ⚠️ Notification2 +method |
|  | `c8y remoteaccess configurations` | — | `— (dynamic)` | M | ⚠️ Remoteaccess? |
|  | `c8y software versions` | — | `— (dynamic)` | M | ⚠️ Repository.Software +method |
|  | `c8y tenants applications` | — | `— (dynamic)` | M | ⚠️ Tenants +method |
|  | `c8y tenants disable` | PUT | `/tenant/tenants/{id}` | M | ⚠️ Tenants +method |
|  | `c8y tenants disableapplication` | DELETE | `/tenant/tenants/{tenant}/applications/{application}` | M | ⚠️ Tenants +method |
|  | `c8y tenants enable` | PUT | `/tenant/tenants/{id}` | M | ⚠️ Tenants +method |
|  | `c8y tenants enableapplication` | POST | `/tenant/tenants/{tenant}/applications` | M | ⚠️ Tenants +method |
|  | `c8y ui plugins` | — | `— (dynamic)` | M | ⚠️ UIPlugins? |
|  | `c8y userreferences addusertogroup` | POST | `/user/{tenant}/groups/{group}/users` | M | ⚠️ Users +method |
|  | `c8y userreferences deleteuserfromgroup` | DELETE | `/user/{tenant}/groups/{group}/users/{user}` | M | ⚠️ Users +method |
|  | `c8y userreferences listgroupmembership` | GET | `/user/{tenant}/groups/{id}/users` | M | ⚠️ Users +method |
|  | `c8y userroles addroletogroup` | POST | `/user/{tenant}/groups/{group}/roles` | M | ⚠️ Userroles +method |
|  | `c8y userroles addroletouser` | POST | `/user/{tenant}/users/{user}/roles` | M | ⚠️ Userroles +method |
|  | `c8y userroles deleterolefromgroup` | DELETE | `/user/{tenant}/groups/{group}/roles/{role}` | M | ⚠️ Userroles +method |
|  | `c8y userroles deleterolefromuser` | DELETE | `/user/{tenant}/users/{user}/roles/{role}` | M | ⚠️ Userroles +method |
|  | `c8y userroles getrolereferencecollectionfromgroup` | GET | `/user/{tenant}/groups/{group}/roles` | M | ⚠️ Userroles +method |
|  | `c8y userroles getrolereferencecollectionfromuser` | GET | `/user/{tenant}/users/{user}/roles` | M | ⚠️ Userroles +method |
|  | `c8y deviceregistration approve` | PUT | `devicecontrol/newDeviceRequests/{id}` | L | ❌ none |
|  | `c8y deviceregistration delete` | DELETE | `devicecontrol/newDeviceRequests/{id}` | L | ❌ none |
|  | `c8y deviceregistration get` | GET | `devicecontrol/newDeviceRequests/{id}` | L | ❌ none |
|  | `c8y deviceregistration getcredentials` | POST | `devicecontrol/deviceCredentials` | L | ❌ none |
|  | `c8y deviceregistration list` | GET | `devicecontrol/newDeviceRequests` | L | ❌ none |
|  | `c8y deviceregistration register` | POST | `devicecontrol/newDeviceRequests` | L | ❌ none |
|  | `c8y events createbinary` | POST | `event/events/{id}/binaries` | L | ⚠️ Events +method |
|  | `c8y events deletebinary` | DELETE | `event/events/{id}/binaries` | L | ⚠️ Events +method |
|  | `c8y events downloadbinary` | GET | `event/events/{id}/binaries` | L | ⚠️ Events +method |
|  | `c8y events updatebinary` | PUT | `event/events/{id}/binaries` | L | ⚠️ Events +method |
|  | `c8y microservices createbinary` | POST | `/application/applications/{id}/binaries` | L | ⚠️ Microservices +method |
