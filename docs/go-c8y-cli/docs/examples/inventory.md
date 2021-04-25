---
layout: default
category: Examples - Powershell
title: Inventory
---


import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Create

### Create a new group, and device and assign the device to the group

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
$Group = c8y devicegroups create --name "AU_Group" -o csv --select id
c8y devices create --name "myDevice01" |
    c8y devicegroups assignDevice --group $Group
```

</TabItem>
<TabItem value="powershell">

```powershell
$Group = c8y devicegroups create --name "AU_Group" | fromjson
c8y devices create --name "myDevice01" |
    c8y devicegroups assignDevice --group $Group.id
```

</TabItem>
</Tabs>


## Update

### Find a managed object by name, and then update it's name

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
c8y inventory find --query "name eq 'AU_Group'" --pageSize 1 |
    c8y inventory update --newName "AUS_Group"
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y inventory find --query "name eq 'AU_Group'" --pageSize 1 |
    c8y inventory update --newName "AUS_Group"
```

</TabItem>
</Tabs>


### Adding update firmware and software capabilities (using supported operations)

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
c8y devices list --query "name eq 'myDevice*'" |
    c8y inventory update \
        --template "{c8y_SupportedOperations: ['c8y_Firmware', 'c8y_SoftwareList']}"
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y devices list --query "name eq 'myDevice*'" |
    c8y inventory update `
        --template "{c8y_SupportedOperations: ['c8y_Firmware', 'c8y_SoftwareList']}"
```

</TabItem>
</Tabs>


### Remove a fragment

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
c8y inventory find --query "name eq 'myDevice*'" |
    c8y inventory update --template "{c8y_Hardware: null}"
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y inventory find --query "name eq 'myDevice*'" |
    c8y inventory update --template "{c8y_Hardware: null}"
```

</TabItem>
</Tabs>


## Child Assets and Devices

### Get a list of child assets of a group

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
c8y devicegroups listAssets --id "AUS_Group"
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y devicegroups listAssets --id "AUS_Group"
```

</TabItem>
</Tabs>
