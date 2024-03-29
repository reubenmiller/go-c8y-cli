---
layout: default
category: Concepts
title: Activity Log
---

import CodeExample from '@site/src/components/CodeExample';

### Overview

The activity log is used as a local protocol to keep track of which commands and requests were sent by go-c8y-cli. It does not store the whole requests and responses (as that would take up too much space), so only meta information about the requests and responses are stored to file.

The activity log was originally created to keep a record of interactions with a production environment, however it has proven useful in all types of Cumulocity tenants (production, quality assurance and development), as it can be really helpful to look at your local interactions with the platform by logging the HTTP status codes along with the response times so you can see how different types commands/requests compare.

The following information is available in the activity log:

* Command
  * Timestamp
  * Command arguments

* Request
  * Accept header
  * Cumulocity Processing Mode
  * HTTP Method
  * Host
  * Path
  * Query parameters

* Response
  * Outgoing timestamp
  * HTTP Status Code
  * Response time (in milliseconds)
  * Response self link (i.e. the `.self` property of the response)

The activity log information is split into different entry types. Each type records slightly different information. The following is a list of the types:

|type|description|
|----|-----------|
|command|go-c8y-cli command used to send the request to|
|request|HTTP request information (specific fields of the request and response)|
|user|user information, i.e. if you don't confirm a prompt etc.|


The request and response are contained in the same entry, however the command has its own entry. The relationship between the go-c8y-cli command and the requests is stored via the `ctx` (context) property, so map commands to their requests.

### Activity log details

The activity log is a json line file (i.e. one object per line). Below is an example of its contents:

```json title="file: ~/.go-c8y-cli/activityLog/c8y.activitylog.2021-04-18.json"
{"time":"2021-04-18T09:46:00.20660075+02:00","ctx":"xtuzTYWn","type":"command","arguments":["devices","list"]}
{"time":"2021-04-18T09:46:00.20737555+02:00","ctx":"RlOXlJdD","type":"command","arguments":["devices","get","--progress","--delay","1000"]}
{"time":"2021-04-18T09:46:00.38602845+02:00","ctx":"xtuzTYWn","type":"request","method":"GET","host":"example.com","path":"/inventory/managedObjects","query":"q=$filter= $orderby=name","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":118,"responseSelf":"https://example.com/inventory/managedObjects?q=$filter%3D%20$orderby%3Dname&pageSize=5&currentPage=1"}
{"time":"2021-04-18T09:46:00.56222815+02:00","ctx":"RlOXlJdD","type":"request","method":"GET","host":"example.com","path":"/inventory/managedObjects/480957","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":99,"responseSelf":"https://example.com/inventory/managedObjects/480957"}
{"time":"2021-04-18T09:46:01.59354455+02:00","ctx":"RlOXlJdD","type":"request","method":"GET","host":"example.com","path":"/inventory/managedObjects/481037","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":29,"responseSelf":"https://example.com/inventory/managedObjects/481037"}
{"time":"2021-04-18T09:46:02.62316225+02:00","ctx":"RlOXlJdD","type":"request","method":"GET","host":"example.com","path":"/inventory/managedObjects/480861","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":21,"responseSelf":"https://example.com/inventory/managedObjects/480861"}
{"time":"2021-04-18T09:46:03.65028965+02:00","ctx":"RlOXlJdD","type":"request","method":"GET","host":"example.com","path":"/inventory/managedObjects/480862","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":20,"responseSelf":"https://example.com/inventory/managedObjects/480862"}
{"time":"2021-04-18T09:46:04.69254065+02:00","ctx":"RlOXlJdD","type":"request","method":"GET","host":"example.com","path":"/inventory/managedObjects/461902","query":"","accept":"application/json","processingMode":"","statusCode":200,"responseTimeMS":35,"responseSelf":"https://example.com/inventory/managedObjects/461902"}
```

The activity logs are all stored in the same folder (regardless of session), however this can be changed for each session by running:

<CodeExample transform="false">

```bash
# change activity log folder
c8y settings update activityLog.path ~/.go-c8y-cli/activitylogs
```

</CodeExample>

The current activity log settings can be displayed by running

<CodeExample transform="false">

```bash
c8y settings list --select activitylog -o json
```

</CodeExample>

```js title="Output"
{
  "activitylog": {
    // currentPath is a read-only setting as it is dynamically generated by go-c8y-cli to make it easier to get the current path
    "currentPath": "/home/vscode/.go-c8y-cli/activityLog/c8y.activitylog.2021-04-18.json",
    "enabled": true,
    "methodfilter": "GET PUT POST DELETE",
    "path": "$C8Y_HOME/activityLog"
  }
}
```

**Note:** `$C8Y_HOME` is configured to `~/.go-c8y-cli` by default, however you can set the environment variable in your shell profile to point to a custom location.

The current activity log file includes the current date timestamp so that the files can be individually managed and delete when they are no longer required. There is currently no automatic rotation of these files, so you will have to delete them manually.

For example, the following one-liner will delete all activity logs older than 14 days.

<CodeExample transform="false">

```bash
find $(dirname $(c8y settings list --select activitylog.currentPath -o csv)) \
  -type f \
  -mtime +14 \
  -name "c8y.activitylog*.json" \
  -delete
```

```powershell
Get-ChildItem -Path (Split-Path (c8y settings list --select activitylog.currentPath -o csv) -Parent) `
  -Recurse `
  -Filter "c8y.activitylog*.json" |
    Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-14)} |
    Remove-Item
```

```powershell
Get-ChildItem -Path (Split-Path (c8y settings list --select activitylog.currentPath -o csv) -Parent) `
  -Recurse `
  -Filter "c8y.activitylog*.json" |
    Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-14)} |
    Remove-Item
```

</CodeExample>

## Querying the activity log

go-c8y-cli also provides commands to display and query the activity log to make it easier to file. If you are using the table view (and have the views configured), then only the request entires will be shown by default.

## Examples

### Get activity log entries

<CodeExample transform="false">

```bash
c8y activitylog list
```

</CodeExample>

````markdown title="Output"
| time                               | ctx           | type         | method     | path                       | query      | statusCode | responseTimeMS |
|------------------------------------|---------------|--------------|------------|----------------------------|------------|------------|----------------|
| 2021-04-18T18:27:19.72950435Z      | nqdfoUex      | request      | GET        | /tenant/currentTenant      |            | 200        | 131            |
| 2021-04-18T18:27:19.95681325Z      | PlJTkVuK      | request      | GET        | /user/currentUser          |            | 200        | 94             |
| 2021-04-18T18:27:20.24930375Z      | kOkZXghv      | request      | POST       | /inventory/managedObjects  |            | 201        | 119            |
| 2021-04-18T18:27:20.60696405Z      | CrAupSgb      | request      | DELETE     | /inventory/managedObjects… |            | 204        | 85             |
| 2021-04-18T18:29:04.47508385Z      | gsILicGq      | request      | GET        | /tenant/currentTenant      |            | 200        | 116            |
| 2021-04-18T18:29:04.68757475Z      | ETWUoSNQ      | request      | GET        | /user/currentUser          |            | 200        | 80             |
| 2021-04-18T18:29:04.93851105Z      | QtEbuAGi      | request      | POST       | /inventory/managedObjects  |            | 201        | 111            |
````

### Get activity log entries from the last 2 hours

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom -2h
```

</CodeExample>

### Get activity log entries for all requests types (as json)

Piping to the cli tool `jq` can be a useful way to display all of the properties (not just those defined in the view)

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom -2h --type all | jq -c
```

</CodeExample>

```json title="Output"
{"arguments":["currentuser","get","--confirm=false"],"ctx":"IiMyVQfk","time":"2021-04-18T18:30:02.78461065Z","type":"command"}
{"accept":"application/json","ctx":"IiMyVQfk","host":"example.com","method":"GET","path":"/user/currentUser","processingMode":"","query":"","responseSelf":"https://t12345.example.com/user/currentUser","responseTimeMS":73,"statusCode":200,"time":"2021-04-18T18:30:02.93165455Z","type":"request"}
{"arguments":["devices","create","--force","--confirm=false","--template","/workspaces/go-c8y-cli/tools/PSc8y/dist/PSc8y/Templates/test.device.jsonnet"],"ctx":"pYPYpXzk","time":"2021-04-18T18:30:03.02579355Z","type":"command"}
{"accept":"application/json","ctx":"pYPYpXzk","host":"example.com","method":"POST","path":"/inventory/managedObjects","processingMode":"","query":"","responseSelf":"https://t12345.example.com/inventory/managedObjects/493890","responseTimeMS":95,"statusCode":201,"time":"2021-04-18T18:30:03.16729655Z","type":"request"}
{"arguments":["alarms","create","--confirm=false","--type=c8y_TestAlarm1","--force","--device=493890","--time=-0s","--text=Test alarm 1","--severity=MAJOR"],"ctx":"eDwpckvV","time":"2021-04-18T18:30:03.22074955Z","type":"command"}
{"accept":"application/json","ctx":"eDwpckvV","host":"example.com","method":"POST","path":"/alarm/alarms","processingMode":"","query":"","responseSelf":"https://t12345.example.com/alarm/alarms/493891","responseTimeMS":109,"statusCode":201,"time":"2021-04-18T18:30:03.36080995Z","type":"request"}
{"arguments":["alarms","create","--confirm=false","--type=c8y_TestAlarm2","--force","--device=493890","--time=-0s","--text=Test alarm 2","--severity=MAJOR"],"ctx":"IaTWjdwa","time":"2021-04-18T18:30:03.41817965Z","type":"command"}
{"accept":"application/json","ctx":"IaTWjdwa","host":"example.com","method":"POST","path":"/alarm/alarms","processingMode":"","query":"","responseSelf":"https://t12345.example.com/alarm/alarms/493893","responseTimeMS":93,"statusCode":201,"time":"2021-04-18T18:30:03.53614835Z","type":"request"}
```

You can also interactive by piping to the linux command `more`

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom -2h --type all -o json -c | more
```

</CodeExample>

### Get requests with a status code of 201 and the path includes "alarms"

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom -3h --type request --filter "statusCode eq 201" --filter "path like *alarms*"
```

</CodeExample>

```markdown title="Output"
| time                               | ctx           | type         | method     | path               | query      | statusCode | responseTimeMS |
|------------------------------------|---------------|--------------|------------|--------------------|------------|------------|----------------|
| 2021-04-18T18:30:03.36080995Z      | eDwpckvV      | request      | POST       | /alarm/alarms      |            | 201        | 109            |
| 2021-04-18T18:30:03.53614835Z      | IaTWjdwa      | request      | POST       | /alarm/alarms      |            | 201        | 93             |
| 2021-04-18T18:30:03.73242335Z      | GlNokisB      | request      | POST       | /alarm/alarms      |            | 201        | 87             |
| 2021-04-18T18:30:03.90711515Z      | AtJiIwth      | request      | POST       | /alarm/alarms      |            | 201        | 95             |
| 2021-04-18T18:30:04.08711435Z      | AQMAbqid      | request      | POST       | /alarm/alarms      |            | 201        | 91             |
| 2021-04-18T18:30:04.27722095Z      | RzNiXVmx      | request      | POST       | /alarm/alarms      |            | 201        | 94             |
```

### Save delete response times for inventory API requests to a csv file

<CodeExample transform="false">

```bash
c8y activitylog list \
  --dateFrom -3h \
  --dateTo -1h \
  --type request \
  --filter "method like delete" \
  --filter "path like **inventory**" \
  --select time,path,method,responseTimeMS -o csvheader > delete_times.csv
```

```powershell
c8y activitylog list \
  --dateFrom -3h \
  --dateTo -1h \
  --type request \
  --filter "method like delete" \
  --filter "path like **inventory**" \
  --select time,path,method,responseTimeMS -o csvheader > delete_times.csv
```

</CodeExample>

```csv title="Output"
time,path,method,responseTimeMS
2021-04-18T18:27:20.60696405Z,/inventory/managedObjects/494211,DELETE,85
2021-04-18T18:29:06.66697205Z,/inventory/managedObjects/493889,DELETE,91
2021-04-18T18:29:06.80089565Z,/inventory/managedObjects/494129,DELETE,89
2021-04-18T18:30:07.17302635Z,/inventory/managedObjects/493890,DELETE,83
2021-04-18T18:31:26.81591365Z,/inventory/managedObjects/494142,DELETE,79
2021-04-18T18:31:33.87323795Z,/inventory/managedObjects/494306,DELETE,116
```

### Get response times of recent api calls

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom "-10min"
```

</CodeExample>

:::note
Only requests for the current session are shown
:::

### Get response times of recent api calls

<CodeExample transform="false">

```bash
c8y activitylog list --dateFrom "-10min" --filter "path like *alarm*" --select time,path,method,responseTimeMS
```

</CodeExample>

```bash title="Output"
| time                               | path               | method     | responseTimeMS |
|------------------------------------|--------------------|------------|----------------|
| 2021-04-18T20:36:16.25192145Z      | /alarm/alarms      | GET        | 186            |
```

:::note
Only requests for the current session are shown
:::

### Print out the response time from all activity logs in csv format

Use the `c8y settings list` to get the directory where the activity log files are stored, then use the linux `find` command to find the activity log json files, and then call `c8y util show` to filter the input json file and 

<CodeExample transform="false">

```bash
find \
  $(dirname $(c8y settings list --select activitylog.currentPath -o csv)) \
  -type f \
  -name "c8y.activity*.json" \
  -exec c8y util show --input {} --filter "type like request*" --select "responseTimeMS" -o csv \;
```

```powershell
Get-ChildItem -Path (Split-Path (c8y settings list --select activitylog.currentPath -o csv) -Parent) `
  -Recurse `
  -Filter "c8y.activitylog*.json" |
    Foreach-Object {
      c8y util show --input $_ --filter "type like request*" --select "responseTimeMS" -o csv
    }
```

```powershell
Get-ChildItem -Path (Split-Path (c8y settings list --select activitylog.currentPath -o csv) -Parent) `
  -Recurse `
  -Filter "c8y.activitylog*.json" |
    Foreach-Object {
      c8y util show --input $_ --filter "type like request*" --select "responseTimeMS" -o csv
    }
```

</CodeExample>

```text title="Output"
202
94
163
45
50
50
46
```
