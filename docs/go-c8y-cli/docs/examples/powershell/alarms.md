---
layout: default
category: Examples - Powershell
title: Alarms
---

### Get

#### Get a list of alarms in the last 30 days

```powershell
Get-AlarmCollection -DateFrom "-30d"
```

**Response**

```plaintext
id    status type     severity text           count source                            creationTime
--    ------ ----     -------- ----           ----- ------                            ------------
68949 ACTIVE testType MAJOR    Custom Event 1 1     @{name=TestDeviceLekJJSsxPT; ...} 01/03/2020 16:41:13
68743 ACTIVE testType MAJOR    Custom Event 1 1     @{name=TestDevicerbROuSshbk; ...} 01/03/2020 16:39:47
67856 ACTIVE testType MAJOR    Custom Event 1 1     @{name=TestDevicebStiOwmjhA; ...} 01/03/2020 16:18:15
67698 ACTIVE testType MAJOR    Custom Event 1 1     @{name=TestDevicenTdCVxkePC; ...} 01/03/2020 16:16:20
65889 ACTIVE testType MAJOR    Custom Event 1 1     @{name=TestDeviceGGMWEMjaEG; ...} 01/03/2020 08:08:00
```

Additional filtering is possible using some parameters like `-DateFrom`, `-DateTo`, `-Fragment`, `-Status`, `-Fragment`, `-Severity` etc. For a full list of parameters use the Powershell help function `help Get-AlarmCollection -Full`

#### Get active alarms for a device by name

```powershell
Get-AlarmCollection -Device device01 -Status ACTIVE
```

**Response**

```plaintext
id    status type     severity text           count source                 creationTime
--    ------ ----     -------- ----           ----- ------                 ------------
68949 ACTIVE testType MAJOR    Custom Event 1 1     @{name=device01; ... } 01/03/2020 16:41:13
```

### Create

#### Create a new alarm for a device

```powershell
New-Alarm `
    -Device "device01" `
    -Time "-0s" `
    -Type "c8y_TestAlarm" `
    -Severity "MAJOR" `
    -Text "Test Alarm"
```

**Response**

```plaintext
id    status type          severity text       count source                creationTime
--    ------ ----          -------- ----       ----- ------                ------------
70657 ACTIVE c8y_TestAlarm MAJOR    Test Alarm 1     @{name=device01; ...} 01/04/2020 14:26:08
```

### Update

#### Update alarm status to ACKNOWLEDGED

```powershell
Update-Alarm -Id 70657 -Status ACKNOWLEDGED
```

**Response**

```plaintext
id    status       type          severity text       count source                     creationTime
--    ------       ----          -------- ----       ----- ------                     ------------
70657 ACKNOWLEDGED c8y_TestAlarm MAJOR    Test Alarm 1     @{name=device01; self=...} 01/04/2020 14:26:08
```

#### Updating multiple alarms

Existing alarms from a device can be updated based on queries. This can be helpful when updating a large number of alarms

```powershell
Update-AlarmCollection -Device device01 -Status ACTIVE -NewStatus ACKNOWLEDGED
```

**Response**

None

### Delete/Remove

#### Remove an alarm

```powershell
Remove-Alarm -Id 70657
```

**Response**

None

### Advanced use cases

#### Subscribing to alarms for a device

```powershell
Watch-Alarm -Device device01
```

**Response**

```plaintext
{"severity":"MAJOR","creationTime":"2020-01-04T14:26:08.989Z","count":3,"history":{"auditRecords":[],"self":"http://goc8yci01.eu-latest.cumulocity.com/audit/auditRecords"},"source":{"self":"http://goc8yci01.eu-latest.cumulocity.com/inventory/managedObjects/70948","id":"70948"},"type":"c8y_TestAlarm","firstOccurrenceTime":"2020-01-04T15:26:08.958+01:00","self":"http://goc8yci01.eu-latest.cumulocity.com/alarm/alarms/70657","time":"2020-01-04T15:51:23.575+01:00","text":"Test Alarm","id":"70657","status":"ACKNOWLEDGED"}

severity            : MAJOR
creationTime        : 01/04/2020 14:26:08
count               : 3
history             : @{auditRecords=System.Object[]; self=http://goc8yci01.eu-latest.cumulocity.com/audit/auditRecords}
source              : @{self=http://goc8yci01.eu-latest.cumulocity.com/inventory/managedObjects/70948; id=70948}
type                : c8y_TestAlarm
firstOccurrenceTime : 01/04/2020 15:26:08
self                : http://goc8yci01.eu-latest.cumulocity.com/alarm/alarms/70657
time                : 01/04/2020 15:51:23
text                : Test Alarm
id                  : 70657
status              : ACKNOWLEDGED
```
