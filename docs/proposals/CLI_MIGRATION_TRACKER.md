# CLI → v2 migration tracker (per-command)

One row per individual go-c8y-cli command still on a spec-generated `*.auto.go`, cross-referenced against the go-c8y **v2** SDK. Fill the **Pri** (priority) column to plan; sorted by effort then SDK-readiness but meant to be re-sorted/filtered freely. Regenerate with `python3 docs/proposals/gen_tracker.py`.

**Remaining commands: 134.**  Effort — S:68 · M:53 · L:13.  SDK call — ✅ ready:68 · ⚠️ service-exists-method-missing:57 · ❌ no-service:9.

**Effort:** `S` ≈ ≤½ day (SDK method ready, mechanical) · `M` ≈ ~1 day (new SDK method / resolution / query-build / sub-resource) · `L` ≈ multi-day (new SDK service, or binary/multipart).

**SDK call:** ✅ `Service.Method` exists · ⚠️ service exists but this method must be added (`Service?` = surface unverified) · ❌ no service at all.

| Pri | Command | HTTP | Endpoint | Effort | SDK call |
|:--:|---|:--:|---|:--:|---|
|  | `c8y applications create` | POST | `/application/applications` | S | ✅ Applications.Create |
|  | `c8y applications delete` | DELETE | `/application/applications/{id}` | S | ✅ Applications.Delete |
|  | `c8y applications get` | GET | `/application/applications/{id}` | S | ✅ Applications.Get |
|  | `c8y applications list` | GET | `/application/applications` | S | ✅ Applications.List |
|  | `c8y applications update` | PUT | `/application/applications/{id}` | S | ✅ Applications.Update |
|  | `c8y configuration create` | POST | `inventory/managedObjects` | S | ✅ Repository.Configuration.Create |
|  | `c8y configuration delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ Repository.Configuration.Delete |
|  | `c8y configuration get` | GET | `inventory/managedObjects/{id}` | S | ✅ Repository.Configuration.Get |
|  | `c8y configuration list` | GET | `inventory/managedObjects` | S | ✅ Repository.Configuration.List |
|  | `c8y configuration update` | PUT | `inventory/managedObjects/{id}` | S | ✅ Repository.Configuration.Update |
|  | `c8y devicegroups create` | POST | `inventory/managedObjects` | S | ✅ Devicegroups.Create |
|  | `c8y devicegroups delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ Devicegroups.Delete |
|  | `c8y devicegroups get` | GET | `inventory/managedObjects/{id}` | S | ✅ Devicegroups.Get |
|  | `c8y devicegroups list` | GET | `inventory/managedObjects` | S | ✅ Devicegroups.List |
|  | `c8y devicegroups listassets` | GET | `inventory/managedObjects/{id}/childAssets` | S | ✅ Devicegroups.List |
|  | `c8y devicegroups update` | PUT | `inventory/managedObjects/{id}` | S | ✅ Devicegroups.Update |
|  | `c8y deviceprofiles create` | POST | `inventory/managedObjects` | S | ✅ ManagedObjects.Create |
|  | `c8y deviceprofiles delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Delete |
|  | `c8y deviceprofiles get` | GET | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Get |
|  | `c8y deviceprofiles list` | GET | `inventory/managedObjects` | S | ✅ ManagedObjects.List |
|  | `c8y deviceprofiles update` | PUT | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Update |
|  | `c8y devices getchild` | GET | `inventory/managedObjects/{device}/childDevices/{reference}` | S | ✅ Devices.Get |
|  | `c8y devices getsupportedmeasurements` | GET | `inventory/managedObjects/{device}/supportedMeasurements` | S | ✅ Devices.ListSupportedMeasurements |
|  | `c8y devices getsupportedseries` | GET | `inventory/managedObjects/{device}/supportedSeries` | S | ✅ Devices.ListSupportedSeries |
|  | `c8y devices listassets` | GET | `inventory/managedObjects/{id}/childAssets` | S | ✅ Devices.List |
|  | `c8y devices listchildren` | GET | `inventory/managedObjects/{device}/childDevices` | S | ✅ Devices.List |
|  | `c8y firmware create` | POST | `inventory/managedObjects` | S | ✅ Repository.Firmware.Create |
|  | `c8y firmware delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ Repository.Firmware.Delete |
|  | `c8y firmware get` | GET | `inventory/managedObjects/{id}` | S | ✅ Repository.Firmware.Get |
|  | `c8y firmware list` | GET | `inventory/managedObjects` | S | ✅ Repository.Firmware.List |
|  | `c8y firmware update` | PUT | `inventory/managedObjects/{id}` | S | ✅ Repository.Firmware.Update |
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
|  | `c8y smartgroups create` | POST | `inventory/managedObjects` | S | ✅ ManagedObjects.Create |
|  | `c8y smartgroups delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Delete |
|  | `c8y smartgroups get` | GET | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Get |
|  | `c8y smartgroups list` | GET | `inventory/managedObjects` | S | ✅ ManagedObjects.List |
|  | `c8y smartgroups update` | PUT | `inventory/managedObjects/{id}` | S | ✅ ManagedObjects.Update |
|  | `c8y software create` | POST | `inventory/managedObjects` | S | ✅ Repository.Software.Create |
|  | `c8y software delete` | DELETE | `inventory/managedObjects/{id}` | S | ✅ Repository.Software.Delete |
|  | `c8y software get` | GET | `inventory/managedObjects/{id}` | S | ✅ Repository.Software.Get |
|  | `c8y software list` | GET | `inventory/managedObjects` | S | ✅ Repository.Software.List |
|  | `c8y software update` | PUT | `inventory/managedObjects/{id}` | S | ✅ Repository.Software.Update |
|  | `c8y usergroups create` | POST | `/user/{tenant}/groups` | S | ✅ Usergroups.Create |
|  | `c8y usergroups delete` | DELETE | `/user/{tenant}/groups/{id}` | S | ✅ Usergroups.Delete |
|  | `c8y usergroups get` | GET | `/user/{tenant}/groups/{id}` | S | ✅ Usergroups.Get |
|  | `c8y usergroups list` | GET | `/user/{tenant}/groups` | S | ✅ Usergroups.List |
|  | `c8y usergroups update` | PUT | `/user/{tenant}/groups/{id}` | S | ✅ Usergroups.Update |
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
|  | `c8y applications copy` | POST | `/application/applications/{id}/clone` | M | ⚠️ Applications +method |
|  | `c8y applications listapplicationbinaries` | GET | `/application/applications/{id}/binaries` | M | ⚠️ Applications +method |
|  | `c8y applications versions` | — | `— (dynamic)` | M | ⚠️ Applications +method |
|  | `c8y bulkoperations listoperations` | GET | `devicecontrol/operations` | M | ⚠️ Bulkoperations +method |
|  | `c8y configuration send` | POST | `devicecontrol/operations` | M | ⚠️ Repository.Configuration +method |
|  | `c8y currentapplication get` | GET | `/application/currentApplication` | M | ❌ none |
|  | `c8y currentapplication listsubscriptions` | GET | `/application/currentApplication/subscriptions` | M | ❌ none |
|  | `c8y currentapplication update` | PUT | `/application/currentApplication` | M | ❌ none |
|  | `c8y devicegroups assigndevice` | POST | `inventory/managedObjects/{id}/childAssets` | M | ⚠️ Devicegroups +method |
|  | `c8y devicegroups assigngroup` | POST | `inventory/managedObjects/{id}/childAssets` | M | ⚠️ Devicegroups +method |
|  | `c8y devicegroups children` | — | `— (dynamic)` | M | ⚠️ Devicegroups +method |
|  | `c8y devicegroups unassigndevice` | DELETE | `inventory/managedObjects/{group}/childAssets/{reference}` | M | ⚠️ Devicegroups +method |
|  | `c8y devicegroups unassigngroup` | DELETE | `inventory/managedObjects/{id}/childAssets/{child}` | M | ⚠️ Devicegroups +method |
|  | `c8y devicemanagement certificates` | — | `— (dynamic)` | M | ⚠️ TrustedCertificates +method |
|  | `c8y devices assignchild` | POST | `inventory/managedObjects/{device}/childDevices` | M | ⚠️ Devices +method |
|  | `c8y devices availability` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices children` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices services` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices statistics` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y devices unassignchild` | DELETE | `inventory/managedObjects/{device}/childDevices/{childDevice}` | M | ⚠️ Devices +method |
|  | `c8y devices user` | — | `— (dynamic)` | M | ⚠️ Devices +method |
|  | `c8y features tenants` | — | `— (dynamic)` | M | ⚠️ Features +method |
|  | `c8y firmware patches` | — | `— (dynamic)` | M | ⚠️ Repository.Firmware +method |
|  | `c8y firmware versions` | — | `— (dynamic)` | M | ⚠️ Repository.Firmware +method |
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
|  | `c8y usergroups getbyname` | GET | `/user/{tenant}/groupByName/{name}` | M | ⚠️ Usergroups +method |
|  | `c8y userreferences addusertogroup` | POST | `/user/{tenant}/groups/{group}/users` | M | ⚠️ Users +method |
|  | `c8y userreferences deleteuserfromgroup` | DELETE | `/user/{tenant}/groups/{group}/users/{user}` | M | ⚠️ Users +method |
|  | `c8y userreferences listgroupmembership` | GET | `/user/{tenant}/groups/{id}/users` | M | ⚠️ Users +method |
|  | `c8y userroles addroletogroup` | POST | `/user/{tenant}/groups/{group}/roles` | M | ⚠️ Userroles +method |
|  | `c8y userroles addroletouser` | POST | `/user/{tenant}/users/{user}/roles` | M | ⚠️ Userroles +method |
|  | `c8y userroles deleterolefromgroup` | DELETE | `/user/{tenant}/groups/{group}/roles/{role}` | M | ⚠️ Userroles +method |
|  | `c8y userroles deleterolefromuser` | DELETE | `/user/{tenant}/users/{user}/roles/{role}` | M | ⚠️ Userroles +method |
|  | `c8y userroles getrolereferencecollectionfromgroup` | GET | `/user/{tenant}/groups/{group}/roles` | M | ⚠️ Userroles +method |
|  | `c8y userroles getrolereferencecollectionfromuser` | GET | `/user/{tenant}/users/{user}/roles` | M | ⚠️ Userroles +method |
|  | `c8y applications createbinary` | POST | `/application/applications/{id}/binaries` | L | ⚠️ Applications +method |
|  | `c8y applications deleteapplicationbinary` | DELETE | `/application/applications/{application}/binaries/{binaryId}` | L | ⚠️ Applications +method |
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
