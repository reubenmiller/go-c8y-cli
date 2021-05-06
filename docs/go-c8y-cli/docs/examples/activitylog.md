---
title: Activity log
---

import CodeExample from '@site/src/components/CodeExample';

## Get

### Get ids of managed objected created from the cli

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom -1h --filter "method eq POST" --filter "path like *inventory*" --select responseSelf -o csv |
    head -1 |
    awk -F/ '{print $NF}'
```

```powershell
c8y activitylog list --dateFrom -1h --filter "method eq POST" --filter "path like *inventory*" --select responseSelf -o csv |
    Select-Object -First 1 |
    Foreach-Object { ($_ -split "/")[-1] }
```

```powershell
c8y activitylog list --dateFrom -1h --filter "method eq POST" --filter "path like *inventory*" --select responseSelf -o csv |
    Select-Object -First 1 |
    Foreach-Object { ($_ -split "/")[-1] }
```

</CodeExample>
