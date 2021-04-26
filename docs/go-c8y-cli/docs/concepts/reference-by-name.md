---
title: Reference By Name
---

import CodeExample from '@site/src/components/CodeExample';

### Accessing devices by name

All devices which require a device id, can also be references by their name (as defined by the `.name` property).

If more than one device has the same name, then only the first result will be matched.

The following shows how the get a list of alarms for a device by only referencing the device by its name:

<CodeExample>

```bash
c8y alarms list --device myDevice
```

</CodeExample>

### Get application by name

Applications can also be referenced by its name, making it easier to use:

<CodeExample>

```bash
c8y applications get --id cockpit
```

</CodeExample>
