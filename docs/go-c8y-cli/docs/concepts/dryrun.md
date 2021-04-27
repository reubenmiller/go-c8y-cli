---
layout: default
category: Concepts
title: Dry run (WhatIf)
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeExample from '@site/src/components/CodeExample';

All commands which send a Cumulocity API request support the dry run mode. A dry run is where the user can check the API call without sending the request to the server. This allows the user to view the request information without having to actually send anything to the server (i.e. checking what `c8y devices create --name mydevice --dry` does without creating a device in the platform).

The dry run mode can be activated by supplying the `dry` parameter to any command.

Previously in go-c8y-cli, dry run mode (formally known as WhatIf mode) was only supported in the PowerShell module, `PSc8y`. In go-c8y-cli v2 onwards, dry run mode is supported natively by the client allowing non-PowerShell users to benefit from it.

## Using dry run mode for documentation

Dry run mode can also be used to generate documentation for individual rest requests to make it easier to share with other users or to document the request in a clear format that can be used to implement in other languages.

The dry run mode supports the following output formats

* markdown
* dump (like a curl request dump)
* json
* curl

The output format can be controlled via the `--dryFormat <markdown|dump|json|curl>` parameter.

Sensitive information can be hidden by changing your session profile settings to hide information such as authorization headers. This can be done by running on of the following commands:

For change it permanently for the current session:

<CodeExample transform="false">

```bash
c8y settings update logger.hideSensitive true
```

</CodeExample>

Or the setting can temporarily activated by setting the environment variable (via update settings command).

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash/zsh', value: 'bash', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
eval $(c8y settings update logger.hideSensitive true --shell auto)
```

</TabItem>
<TabItem value="fish">

```bash
c8y settings update logger.hideSensitive true --shell auto | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y settings update logger.hideSensitive true --shell auto | Out-String | Invoke-Expression
```

</TabItem>
</Tabs>

The following show examples of the output from each of the formats.

### markdown (default)

Default output mode which is useful because it can easily be copy/pasted into a markdown document or directing into your online documentations website (i.e. Confluence or similar products). The url path is shown without the server address so it can be kept generic, though the full path is also shown above it in case you need the full path.

<CodeExample>

```bash
c8y devices create --name "device001" --dry --dryFormat markdown
```

</CodeExample>

````markdown title="Output"
What If: Sending [POST] request to [https://example.com/inventory/managedObjects]

### POST /inventory/managedObjects

| header            | value
|-------------------|---------------------------
| Accept            | application/json 
| Authorization     | Basic  {base64 tenant/username:password}
| Content-Type      | application/json 

#### Body

```json
{
  "c8y_IsDevice": {},
  "name": "device001"
}
```
````

### dump (http output)

The dump format shows the http request in its HTTP/1.x wire representation, the same format shown by curl when using the verbose parameter. This output is useful if you want a plain text representation of the request.

<CodeExample>

```bash
c8y devices create --name "device001" --dry --dryFormat dump
```

</CodeExample>

```text title="Output"
POST /inventory/managedObjects HTTP/1.1
Host: example.com
Accept: application/json
Authorization: Basic {base64 tenant/username:password}
Content-Type: application/json
User-Agent: go-client
X-Application: go-client

{"c8y_IsDevice":{},"name":"device001"}
```

### json

The json output is useful to retrieve request information back easy-to-parse format, so the fields can be extracted using `jq` or another other json parsing tool.

<CodeExample>

```bash
c8y devices create --name "device001" --dry --dryFormat json

# just get the full url for the request by piping the output to jq
c8y devices create --name "device001" --dry --dryFormat json | jq -r '.url'
```

</CodeExample>

```json title="Output"
{
  "url": "https://example.com/inventory/managedObjects",
  "host": "https://example.com",
  "pathEncoded": "/inventory/managedObjects",
  "path": "/inventory/managedObjects",
  "method": "POST",
  "headers": {
    "Accept": "application/json",
    "Authorization": "Basic {base64 tenant/username:password}",
    "Content-Type": "application/json"
  },
  "body": {
    "c8y_IsDevice": {},
    "name": "device001"
  },
  "shell": "curl -X 'POST' -d '{\"c8y_IsDevice\":{},\"name\":\"device001\"}' -H 'Accept: application/json' -H 'Authorization: Basic {base64 tenant/username:password}' -H 'Content-Type: application/json' 'https://example.com/inventory/managedObjects'",
  "powershell": "curl -X 'POST' -d '{\\\"c8y_IsDevice\\\":{},\\\"name\\\":\\\"device001\\\"}' -H 'Accept: application/json' -H 'Authorization: Basic {base64 tenant/username:password}' -H 'Content-Type: application/json' 'https://example.com/inventory/managedObjects'"
}
```

### curl (beta)

The curl output shows the equivalent curl command that can be used to send the same request. The curl commands are printed in markdown one showing for shell and one for PowerShell (due to different string quoting conventions).


:::caution

If the command is using a multipart/form-data upload (i.e. a file upload), then the curl command might not be 100% correct, and might need a minor modification to get it to work.
:::

<CodeExample>

```bash
c8y devices create --name "device001" --dry --dryFormat curl
```

</CodeExample>

````markdown title="Output"
##### Curl (shell)

```bash
curl -X 'POST' -d '{"c8y_IsDevice":{},"name":"device001"}' -H 'Accept: application/json' -H 'Authorization: Basic {base64 tenant/username:password}' -H 'Content-Type: application/json' 'https://example.com/inventory/managedObjects'
```

##### Curl (PowerShell)

```powershell
curl -X 'POST' -d '{\"c8y_IsDevice\":{},\"name\":\"device001\"}' -H 'Accept: application/json' -H 'Authorization: Basic {base64 tenant/username:password}' -H 'Content-Type: application/json' 'https://example.com/inventory/managedObjects'
```
````

## Example: Write request details to file

The output dry run output is printed to standard output, and can be redirected to file or piped to a downstream tool.

<CodeExample>

```bash
c8y alarms list \
  --dateFrom "-2h" \
  --severity CRITICAL \
  --status ACTIVE \
  --dry --dryFormat markdown > list_critical_alarms.md
```

</CodeExample>

````markdown title="file: list_critical_alarms.md"
What If: Sending [GET] request to [https://example.com/alarm/alarms?dateFrom=2021-04-18T16:54:45.36245505Z&severity=CRITICAL&status=ACTIVE]

### GET /alarm/alarms?dateFrom=2021-04-18T16:54:45.36245505Z&severity=CRITICAL&status=ACTIVE

| header            | value
|-------------------|---------------------------
| Accept            | application/json 
| Authorization     | Basic {base64 tenant/username:password}
````
