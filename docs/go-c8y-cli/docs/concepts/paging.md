---
layout: default
category: Concepts
title: Paging
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

### Limiting query results

#### Background

Cumulocity uses paging for any query which returns a collection of resources. For example, when getting a list of devices (managed objects), the query will limit the number of devices returns to the defined `pageSize`.

The `pageSize` has an upper limit of 2000 which is enforced on the server side.

In addition, Cumulocity also supports a `currentPage` parameter which can be used to control which page from the collection the server should return. The `currentPage` defaults 1, however can be changed to return a different page of the collection.

Please read the [Cumulocity IoT documentation](http://cumulocity.com/guides/reference/rest-implementation/#rest-usage) for further details about paging.


#### Using paging on the command line

The c8y cli tool supports interacting with Cumulocity's paging, and provides a few enhancements to help to get all of the data that you need to be productive.

The following command line arguments are supported by all collection related commands.

| PSc8y | c8y | Description |
|-------|---------|---------|
| `-PageSize` | `--pageSize` | Set the maximum number of results included in the response. Limited to 2000 by the server. |
| `-CurrentPage` | `--currentPage` | Page number (or slice) of the results which should be returned. Defaults to 1. |
| `-TotalPages` | `--totalPages` | Total number of pages to return. The pages will be collected serially starting from the `currentPage`  |
| `-IncludeAll` | `--includeAll` | Include all pages from the collection. IncludeAll will automatically set the pageSize to the maximum value `2000` regardless what value was given to `pageSize` |


A few examples will now be detailed to clarify the usage of the parameters in real life scenarios.

#### Example: Get all devices with a matching name prefix

If you have a large number of devices and you want to retrieve all of the results for some offline analysis.

The `includeAll` parameter is used to get all of the results.

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
c8y devices list --name "MyDevices*" --includeAll --select "id,name,*.serialNumber" --output csvheader > devicelist.csv
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-DeviceCollection -Name "MyDevices*" -IncludeAll --select "id,name,*.serialNumber" -Output csvheader > devicelist.csv
```

</TabItem>
</Tabs>


#### Example: Get total number of devices

The total number of devices can be returned by using the technic of setting the `pageSize` to 1, and using the `withTotalPages` parameter. The result will then contain the total number (in the `.statistics.totalPages` property) of whatever entity you have requested. A view has been added to only display the `statistics` fragment by default.

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
c8y devices list --withTotalPages --pageSize 1
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-DeviceCollection -WithTotalPages -PageSize 1
```

</TabItem>
</Tabs>


```json title="output"
| totalPages | pageSize   | currentPage |
|------------|------------|-------------|
| 159        | 1          | 1           |
```

or you can get the raw json response by adding the `raw` parameter


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
c8y devices list --withTotalPages --pageSize 1 --raw
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-DeviceCollection -WithTotalPages -PageSize 1 -Raw
```

</TabItem>
</Tabs>

```json title="output"
{
  "managedObjects": [
    {
      "additionParents": {
        "references": [],
        "self": "https://t12345.example.com/inventory/managedObjects/494210/additionParents"
      },
      "assetParents": {
        "references": [],
        "self": "https://t12345.example.com/inventory/managedObjects/494210/assetParents"
      },
      "c8y_IsDevice": {},
      "childAdditions": {
        "references": [],
        "self": "https://t12345.example.com/inventory/managedObjects/494210/childAdditions"
      },
      "childAssets": {
        "references": [],
        "self": "https://t12345.example.com/inventory/managedObjects/494210/childAssets"
      },
      "childDevices": {
        "references": [],
        "self": "https://t12345.example.com/inventory/managedObjects/494210/childDevices"
      },
      "creationTime": "2021-04-18T18:10:12.940Z",
      "deviceParents": {
        "references": [],
        "self": "https://t12345.example.com/inventory/managedObjects/494210/deviceParents"
      },
      "id": "494210",
      "lastUpdated": "2021-04-18T18:10:12.940Z",
      "name": "device001",
      "owner": "ciuser01",
      "self": "https://t12345.example.com/inventory/managedObjects/494210"
    }
  ],
  "next": "https://t12345.example.com/inventory/managedObjects?q=$filter%3D%20$orderby%3Dname&pageSize=1&currentPage=2&withTotalPages=true",
  "self": "https://t12345.example.com/inventory/managedObjects?q=$filter%3D%20$orderby%3Dname&pageSize=1&currentPage=1&withTotalPages=true",
  "statistics": {
    "currentPage": 1,
    "pageSize": 1,
    "totalPages": 159
  }
}
```


#### Iterating through the results

Since pipeline data is supported natively by go-c8y-cli, any pages results can be efficiently piped to downstream commands.

Instead of retrieving all of the devices at once, you can run a task on each paging result, and then move on to the next page, until all of the paging have been processed.

This has the advantages that all of the results do not need to be kept in memory.


#### Example: Run a custom shell command on each of the results

TODO

**Note:** when piping json it is necessary to escape the double quotes before passing it down the pipeline, this can be done by using `sed`.

```bash
c8y devices list -p 10 | sed 's/"/\\"/g' | xargs -0 -I {} bash -c "echo \"{}\" | jq -r '.name'"
```

Alternatively, the gnu command `parallel` can be used as it handles json from standard input in a more convenient way.

```bash
c8y devices list | parallel --tag echo {} | jq -r '.name'
```

```bash title="output"
device_0001
device_0002
device_0003
device_0004
device_0005
```

#### Example: Add custom shell command in the pipeline

TODO

**Example: Add a fragment to each device**

The following shows how to add a fragment `myNewFragment` to each devices where the name starts with "My".


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
c8y devices list --name "My*" --includeAll |
  c8y devices update --data "myNewFragment.fragmentCreationTime=$( date --iso-8601=seconds )"

# or using templates
c8y devices list --name "My*" --includeAll |
  c8y devices update --template "{ myNewFragment: {fragmentCreationTime: _.Now('0s') }}"
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-DeviceCollection -Name "My*" -IncludeAll |
    Update-ManagedObject -Data "myNewFragment.fragmentCreationTime=$(Format-Date)"

# or using templates
Get-DeviceCollection -Name "My*" -IncludeAll |
    Update-ManagedObject -Template "{ myNewFragment: {fragmentCreationTime: _.Now('0s') }}"
```

</TabItem>
</Tabs>


### Setting a default pageSize

The default pageSize can be controlled via the session or `settings` file or in your session file.

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
c8y settings update defaults.pageSize 20
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y settings update defaults.pageSize 20
```

</TabItem>
</Tabs>
