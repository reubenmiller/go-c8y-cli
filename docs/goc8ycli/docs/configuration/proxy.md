---
title: Proxy
---

import CodeExample from '@site/src/components/CodeExample';

c8y has first class support for corporate proxies. The proxy can either be configured globally by using environment variables, or by using parameters.

### Setting proxy using environment variables

<CodeExample>

```bash
export HTTP_PROXY=http://10.0.0.1:8080
export NO_PROXY=localhost,127.0.0.1,edge01.server
```

```powershell
$env:HTTP_PROXY = "http://10.0.0.1:8080"
$env:NO_PROXY = "localhost,127.0.0.1,edge01.server"
```

</CodeExample>

### Overriding proxy settings for individual commands

The proxy environment variables can be ignore for individual commands by using the `noProxy` parameter.

<CodeExample>

```bash
c8y devices list --noProxy
```

</CodeExample>


Or, the proxy settings can be set for a single command by using the  `proxy` parameter:

<CodeExample>

```bash
c8y devices list --proxy "http://10.0.0.1:8080"
```

</CodeExample>
