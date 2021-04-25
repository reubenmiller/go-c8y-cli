---
title: Alarms
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Get

### Get a list of alarms in the last 30 days


<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms list --dateFrom "-30d"
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-AlarmCollection -DateFrom "-30d"
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id          | status      | type                         | severity   | count      | source.id   | source.name             | creationTime                  | text                                   |
|-------------|-------------|------------------------------|------------|------------|-------------|-------------------------|-------------------------------|----------------------------------------|
| 497826      | ACTIVE      | c8y_UnavailabilityAlarm      | MAJOR      | 1          | 497719      | mobile-device_0005      | 2021-04-24T19:48:47.790Z      | No data received from device within r… |
| 497824      | ACTIVE      | c8y_UnavailabilityAlarm      | MAJOR      | 1          | 497718      | mobile-device_0002      | 2021-04-24T19:48:47.768Z      | No data received from device within r… |
| 497921      | ACTIVE      | c8y_UnavailabilityAlarm      | MAJOR      | 1          | 497914      | mobile-device_0004      | 2021-04-24T19:48:43.796Z      | No data received from device within r… |
```

:::note
Additional filtering is possible using some parameters like `dateFrom`, `dateTo`, `fragmentType`, `status`, `severity` etc. For a full list of parameters use the inbuilt help

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms list --help
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-Help Get-AlarmCollection -Full
```

</TabItem>
</Tabs>
:::

### Get active alarms for a device by name

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms list --device device01 --status ACTIVE
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-AlarmCollection -Device device01 -Status ACTIVE
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id          | status      | type                  | severity   | count      | source.id   | source.name   | creationTime                  | text          |
|-------------|-------------|-----------------------|------------|------------|-------------|---------------|-------------------------------|---------------|
| 497924      | ACTIVE      | c8y_temperature0      | MAJOR      | 3          | 497835      | device01      | 2021-04-25T12:18:00.290Z      | Too cold      |
| 497727      | ACTIVE      | c8y_temperature2      | MINOR      | 4          | 497835      | device01      | 2021-04-25T12:17:46.606Z      | Unknown error |
| 497734      | ACTIVE      | c8y_sensor2           | MINOR      | 2          | 497835      | device01      | 2021-04-25T12:18:54.500Z      | Disconnected  |
| 497725      | ACTIVE      | c8y_sensor0           | MAJOR      | 3          | 497835      | device01      | 2021-04-25T12:17:35.997Z      | Too hot       |
| 497729      | ACTIVE      | c8y_sensor3           | WARNING    | 4          | 497835      | device01      | 2021-04-25T12:17:53.808Z      | Too cold      |
```

## Create

### Create a new alarm for a device

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms create \
    --device "device01" \
    --time "0s" \
    --type "c8y_TestAlarm" \
    --severity "MAJOR" \
    --text "Test Alarm"
```

</TabItem>
<TabItem value="powershell">

```powershell
New-Alarm `
    -Device "device01" `
    -Time "0s" `
    -Type "c8y_TestAlarm" `
    -Severity "MAJOR" `
    -Text "Test Alarm"
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id          | status      | type               | severity   | count      | source.id   | source.name   | creationTime                  | text            |
|-------------|-------------|--------------------|------------|------------|-------------|---------------|-------------------------------|-----------------|
| 497740      | ACTIVE      | c8y_TestAlarm      | MAJOR      | 1          | 497835      | device01      | 2021-04-25T12:22:53.218Z      | Test Alarm      |
```

### Update

### Update alarm status to ACKNOWLEDGED

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms update --id 497740 --status ACKNOWLEDGED
```

</TabItem>
<TabItem value="powershell">

```powershell
Update-Alarm -Id 497740 -Status ACKNOWLEDGED
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id          | status            | type               | severity   | count      | source.id   | source.name   | creationTime                  | text            |
|-------------|-------------------|--------------------|------------|------------|-------------|---------------|-------------------------------|-----------------|
| 497740      | ACKNOWLEDGED      | c8y_TestAlarm      | MAJOR      | 1          | 497835      | device01      | 2021-04-25T12:22:53.218Z      | Test Alarm      |
```

### Updating multiple alarms

Existing alarms from a device can be updated based on queries. This can be helpful when updating a large number of alarms

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms updateCollection --device device01 --status ACTIVE --newStatus ACKNOWLEDGED
```

</TabItem>
<TabItem value="powershell">

```powershell
Update-AlarmCollection -Device device01 -Status ACTIVE -NewStatus ACKNOWLEDGED
```

</TabItem>
</Tabs>


```plaintext title="No output"
```

## Delete/Remove

### Remove an alarm

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms deleteCollection --device device01 --status ACKNOWLEDGED --dateFrom -1d
```

</TabItem>
<TabItem value="powershell">

```powershell
Remove-AlarmCollection -Device device01 -Status ACKNOWLEDGED -DateFrom -1d
```

</TabItem>
</Tabs>

```plaintext title="Output (standard error)"
✓ Deleted /alarm/alarms => 204 No Content
```

## Advanced use cases

### Subscribing to alarms for a device for 180 seconds

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y alarms subscribe --device device01 --duration 180
```

</TabItem>
<TabItem value="powershell">

```powershell
Watch-Alarm -Device device01 -Duration 180
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id          | status      | type             | severity   | count      | source.id   | source.name   | creationTime                  | text               |
|-------------|-------------|------------------|------------|------------|-------------|---------------|-------------------------------|--------------------|
| 497751      | ACTIVE      | c8y_sensor0      | MAJOR      | 1          | 497835      | device01      | 2021-04-25T12:33:28.231Z      | Unknown error      |
```
