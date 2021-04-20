---
layout: default
category: Examples - Powershell
title: Audit Records
---

### Get

#### Retrieve the audit records related to an alarm

The history of an alarm can be recalled by using the following: 

```powershell
Get-Alarm 70657 | Get-AuditRecordCollection
```

**Response**

```plaintext
severity activity      creationTime        source                type  self
-------- --------      ------------        ------                ----  ----
MAJOR    Alarm created 01/04/2020 14:26:09 @{self=...; id=70657} Alarm https://...
MAJOR    Alarm updated 01/04/2020 14:30:05 @{self=...; id=70657} Alarm https://...
```

#### Retrieve the audit records related to an operation

The history of an operation can be recalled by using the following:

```powershell
Get-Operation -Id 70953 | Get-AuditRecordCollection
```

**Response**

```plaintext
id    time                type      activity          text                                                                                          user          source
--    ----                ----      --------          ----                                                                                          ----          ------
70954 01/04/2020 15:13:32 Operation Operation created Operation created: status='PENDING', description='Test operation', device name='device01'.    citest-pwsh01 @{self=https://goc8yci01.eu-latest.cumul…
70955 01/04/2020 15:14:12 Operation Operation updated Operation updated: status='EXECUTING', description='Test operation', device name='device01'.  citest-pwsh01 @{self=https://goc8yci01.eu-latest.cumul…
70956 01/04/2020 15:14:41 Operation Operation updated Operation updated: status='SUCCESSFUL', description='Test operation', device name='device01'. citest-pwsh01 @{self=https://goc8yci01.eu-latest.cumul…
```


### Create

#### Create audit record related to a device

```powershell
New-AuditRecord `
    -Type "ManagedObject" `
    -Time "0s" `
    -Text "Managed Object updated: my_Prop: value" `
    -Source "70948" `
    -Activity "Managed Object updated" `
    -Severity "information"
```

**Response**

```plaintext
id    time                type          activity               text                                   user          source
--    ----                ----          --------               ----                                   ----          ------
70958 01/04/2020 16:20:31 ManagedObject Managed Object updated Managed Object updated: my_Prop: value citest-pwsh01 @{self=https://goc8yci01.eu-latest.cumulocity.com/inventory/managedObjects/70948; id=7…
```
