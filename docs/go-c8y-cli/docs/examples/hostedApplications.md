---
title: Applications - Web (Hosted)
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Hosted (web) Applications


### Create or update a hosted application


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
# Create from a build folder
c8y applications createHostedApplication --name "myapp" --file "./build"

# From a zip file (name will default to the zip filename)
c8y applications createHostedApplication --file "./build/myapp.zip"
```

</TabItem>
<TabItem value="powershell">

```powershell
# Create from a build folder
New-HostedApplication -Name "myapp" -File "./build"

# From a zip file (name will default to the zip filename)
New-HostedApplication -File "./build/myapp.zip"
```

</TabItem>
</Tabs>

:::info
The command will check if the application already exists or not, and create the application if necessary. Otherwise it just uploads the binary and activates it.

If you don't want to activate the binary use the `skipActivation` parameter
:::
