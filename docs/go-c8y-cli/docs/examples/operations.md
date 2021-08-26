---
title: Operations
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Advanced

### Simulate an agent processing operations

Simulate the transition of an operation on a device by moving it from the following states.

* `PENDING` -> `EXECUTING` (after 5 seconds)
* `EXECUTING` -> `SUCCESSFUL` or `FAILED` (randomly chosen) (after 5 seconds)

<CodeExample transform="false">

```bash
c8y operations subscribe --duration 3m --filter "status notlike SUCCESSFUL|FAILED" \
| c8y operations update --delay 5s --force \
    --template "{status: if input.value.status == 'PENDING' then 'EXECUTING' else ['SUCCESSFUL', 'FAILED'][_.Int(0,1)]}"
```

</CodeExample>

Alternatively it could be simplified, however only the output from the last command will be visible on the console.

<CodeExample transform="false">

```bash
c8y operations subscribe --duration 3m --actionTypes CREATE \
| c8y operations update --delay 5s --status EXECUTING --force \
| c8y operations update --delay 5s --status SUCCESSFUL --force
```

</CodeExample>
