---
id: cli
title: go-c8y-cli commands
---

import CodeExample from '@site/src/components/CodeExample';

The following pages are the doc pages for both the `c8y` native command and the PowerShell Module wrapper `PSc8y`. The online docs make it easier to have a look around to see all the commands that are available.

All of the documentation published here is also available from the command line using the following commands

<CodeExample>

```bash
# Show commands
c8y help

# Get help for single command
c8y devices list --help | more
```


```powershell
# Show commands
Get-Command -Module PSc8y

# Get help for single command
Get-Help Get-DeviceCollection -Full
```

</CodeExample>
