---
layout: default
category: Examples - Powershell
title: Device Control
---

### Create

#### Create Bulk operation using custom json

```powershell
$prop = @"
{
 "operationPrototype":{
   "description": "Update firmware to: rmi_base (version: v0.0.1).",
   "c8y_Firmware": {
    "name": "rmi_base",
    "version": "v0.0.1",
    "url": "https://myTenant.eu-latest.cumulocity.com/inventory/binaries/12345"
  },
 },
 "creationRamp":45,
 "groupId":"112011",
 "startDate":"2020-02-01T22:21:22"
}
"@

Invoke-CumulocityRequest -Uri "/devicecontrol/bulkoperations" -Method Post -Data $prop
```
