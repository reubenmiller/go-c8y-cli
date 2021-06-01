---
title: Known limitations
---

import CodeExample from '@site/src/components/CodeExample';

## Large numbers using exponential notation in jsonnet templates

Usage of large numbers (using exponential notation) in templates can cause the server to respond with a 422 status code. This is due to a current limitation in the jsonnet template library.

As a workaround, use the `data` parameter when working with large numbers (and don't specify a `template` parameter). This will preserve the number format as provided.

**Example**

```json title="file: datapoint.largeInt.json"
{
    "c8y_Kpi": {
        "max": 19.1010101E19,
        "description": "Counter"
    }
}
```

<CodeExample>

```sh
c8y inventory create --data "./datapoint.largeInt.json"
```

</CodeExample>

## Using global --session parameter with Two Factor Authentication

go-c8y-cli supports Cumulocity tenants which are configured with Two Factor Authentication (TFA). When activating the session you will be prompted for your TFA code. Once the TFA code has been verified, the token returned by Cumulocity will be saved into the session file. The subsequent command will use the token for all authentication.

Before such a TFA enabled tenant can be referenced via the global `--session <session>` parameter, it needs to be manually activated so that the token is stored inside the related session file.

The following shows how just manually activating the session first can solve this problem.

<CodeExample>

```sh
set-session tenantA
set-session tenantB
c8y devices list    # uses tenantB
c8y devices list --session tenantA  # uses tenantA for a one-off command
```

</CodeExample>

:::note
The token will expire after a period of time (depends on tenant configuration), so if you are still having problems with a session, try to manually activate it first, then retry with the `--session` parameter.
:::
