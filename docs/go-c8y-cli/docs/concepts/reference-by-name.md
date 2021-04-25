---
title: Reference By Name
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

### Accessing devices by name

All devices which require a device id, can also be references by their name (as defined by the `.name` property).

If more than one device has the same name, then only the first result will be matched.

The following shows how the get a list of alarms for a device by only referencing the device by its name:

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
c8y alarms list --device myDevice
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-AlarmCollection -Device myDevice
```

</TabItem>
</Tabs>


### Get application by name

Applications can also be referenced by its name, making it easier to use:

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
c8y applications get --id cockpit
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-Application -Id cockpit
```

</TabItem>
</Tabs>
