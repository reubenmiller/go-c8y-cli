---
title: Proxy
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

c8y has first class support for corporate proxies. The proxy can either be configured globally by using environment variables, or by using parameters.

### Setting proxy using environment variables

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
export HTTP_PROXY=http://10.0.0.1:8080
export NO_PROXY=localhost,127.0.0.1,edge01.server
```

</TabItem>
<TabItem value="powershell">

```powershell
$env:HTTP_PROXY = "http://10.0.0.1:8080"
$env:NO_PROXY = "localhost,127.0.0.1,edge01.server"
```

</TabItem>
</Tabs>

### Overriding proxy settings for individual commands

The proxy environment variables can be ignore for individual commands by using the `noProxy` parameter.

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
c8y devices list --noProxy
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-DeviceCollection -NoProxy
```

</TabItem>
</Tabs>


Or, the proxy settings can be set for a single command by using the  `proxy` parameter:

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
c8y devices list --proxy "http://10.0.0.1:8080"
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-DeviceCollection -Proxy "http://10.0.0.1:8080"
```

</TabItem>
</Tabs>
