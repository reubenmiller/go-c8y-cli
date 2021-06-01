---
title: Activity log
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Demo

<Video
  videoSrcURL="https://asciinema.org/a/414235/embed?speed=1.0&autoplay=false&size=small&rows=30"
  videoTitle="Activitylog example"
  width="90%"
  height="600px"
  ></Video>


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
