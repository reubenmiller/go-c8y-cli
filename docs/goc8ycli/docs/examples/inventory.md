---
layout: default
category: Examples - Powershell
title: Inventory
---

import CodeExample from '@site/src/components/CodeExample';

## Create

### Create a new group, and device and assign the device to the group


<CodeExample>

```bash
$Group = c8y devicegroups create --name "AU_Group" -o csv --select id
c8y devices create --name "myDevice01" |
    c8y devicegroups assignDevice --group $Group
```

```powershell
$Group = c8y devicegroups create --name "AU_Group" | fromjson
c8y devices create --name "myDevice01" |
    c8y devicegroups assignDevice --group $Group.id
```

</CodeExample>

## Update

### Find a managed object by name, and then update it's name

<CodeExample>

```bash
c8y inventory find --query "name eq 'AU_Group'" --pageSize 1 |
    c8y inventory update --newName "AUS_Group"
```

</CodeExample>

### Adding update firmware and software capabilities (using supported operations)


<CodeExample>

```bash
c8y devices list --query "name eq 'myDevice*'" |
    c8y inventory update \
        --template "{c8y_SupportedOperations: ['c8y_Firmware', 'c8y_SoftwareList']}"
```

</CodeExample>

### Remove a fragment

<CodeExample>

```bash
c8y inventory find --query "name eq 'myDevice*'" |
    c8y inventory update --template "{c8y_Hardware: null}"
```

</CodeExample>

## Child Assets and Devices

### Get a list of child assets of a group

<CodeExample>

```bash
c8y devicegroups listAssets --id "AUS_Group"
```

</CodeExample>
