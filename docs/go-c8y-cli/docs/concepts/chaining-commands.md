---
title: Chaining commands
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeExample from '@site/src/components/CodeExample';

Most commands in go-c8y-cli support piped input (i.e. input from standard input (stdin)). This makes it easy to chain commands together.

Previously in go-c8y-cli version 1, the PowerShell wrapper of go-c8y-cli was way to utilize piped data for passing data to downstream commands, for example:

<CodeExample>

```bash
c8y devices list | c8y devices update --data "myValue=1"
```

</CodeExample>

With the go-c8y-cli version 2 release, pipeline input is now supported natively by the c8y binary. Taking the PowerShell example above, the same can be written in any shell (event PowerShell):

By supporting pipeline input, it enables greater flexibility as most of the commands outputs can be feed directly into another command without needing to save the output to a variable first, and in addition, it made it possible to support the following new features/optimizations:

* Concurrency (sending multiple requests at the same time) (via the `workers <int>` parameter)
* Progress indicators (via the `progress` parameter)
* Reduction of calls to c8y binary (as the binary is only called once instead of once per input value)
* Easy to feed data from files (i.e. list of ids or names)

### Piping basics

Let's say you want to lookup an inventory managed object using its id (because you copied it from the browser's url from one of Cumulocity's in-built applications). Normally if you are working with just a single id then you would use the following command:

<CodeExample>

```bash
c8y devices get --id 1234
```

</CodeExample>

Now the same can be written using piped input by using the shell's `echo` command as the input data will be mapped to the `--id` parameter automatically. The parameter which will receive pipeline input is marked in the help with `(accepts pipeline)`, but generally there should mostly be the obvious choice which parameter would make most sense (i.e. `--id`, `--device` etc.). A commands help text can be displayed by adding `--help` to a command, i.e. `c8y devices get --help`.

<CodeExample>

```bash
echo "1234" | c8y devices get
```

```powershell
"1234" | batch | Get-Device
```

</CodeExample>

Now that does not seem so interesting as we are still fetching a single device and that is something Postman can do easily. Let's extend the command by now fetching 2 devices. Each id must be on a new line (so that it is treated as an independent value), so `echo -e` will be used so that we can use the newline character `\n` to separate the two ids.

<CodeExample>

```bash
echo -e "1234\n5678" | c8y devices get
```

```powershell
1234, 5678 | batch | Get-Device
```
</CodeExample>

Again not super interesting as who really wants to type multiple ids in every time you want to get the current values. So instead of echoing the ids, let's store them to file called `ids.txt` (again one id per line):

```text title="file: ids.txt"
1234
5678
```

c8y won't read the file directly, however you can use `cat` to read the file and it will pass contents of the file (line by line) to go-c8y-cli via the pipeline. By using a file it makes it easy to extend the list of ids by just appending new ids to it, and the command stays the same.

<CodeExample>

```bash
cat ids.txt | c8y devices get
```

```powershell
Get-Content ids.txt | batch | Get-Device
```

</CodeExample>

Though finding ids is annoying, so let's find the ids first and then store them to file, and this will act as our "watch list" that we want to use to occasionally fetch the updated information about these devices.

So we can build the input `ids.txt` file by doing a query by type and a custom fragment. We only want ot store the id, name and type of the devices so that the file remains relatively small, however the user has the name and type field to help identify which device the id is representing.

<CodeExample>

```bash
c8y devices list --type "myType*" --query "has(myCustomFragment)" --select id,name,type --output csv > ids.txt
```

```powershell
Get-DeviceCollection -Type "myType*" -Query "has(myCustomFragment)" -Select id,name,type -Output csv > ids.txt
```

</CodeExample>

```csv title="file: ids.txt"
1234,My Device 1
5678,My Device 2
```

This file can be piped in exactly the same way as go-c8y-cli is smart enough to detect that a csv file is given, and it only takes the first column of each row, using a `,` (comma) delimiter.

Now saving ids to file is not always the way to go. It might be ok if you are dealing with devices, however for other Cumulocity entities, such as alarms, operations, events) it does not make too much sense as these entities are only interesting for a particular time period. So let's apply the same piping principles to handling operations.

For example, you are using Cumulocity and you notice there are a lot of pending operations for various devices in the platform. Most likely this occurred because the devices are offline for a longer period, or maybe the device agent does not recognize the operation type and left them in pending instead of cancelling them. You want to cancel the operations that are older than 14 days, however you don't want to delete them as you still want them to keep them for some following up analysis (deleting would be more efficient as there is a delete collection api, however you will lose the operation instead of the status just being set to `FAILED`).

Saving the ids of the operations in the case does not make sense because you don't really care about the operation ids, you only care if the operation is in `PENDING` and older than 14 days. So instead of saving the ids to file, you can just pipe the results from your operations query directly do the update operation command.

<CodeExample>

```bash
c8y operations list --status "PENDING" --dateTo "-14d" --includeAll |  # get list of operations
    c8y operations update \
        --status "FAILED" \
        --failureReason "Cancelled stale operation as it was older than 14 days"
```

```powershell
Get-OperationCollection -Status "PENDING" -DateTo "-14d" -IncludeAll |  # get list of operations
    Update-Operation `
        -Status "FAILED" `
        -FailureReason "Cancelled stale operation as it was older than 14 days"
```

</CodeExample>


The first command `c8y operations list` is responsible for getting the list of operations matching your search criteria. You can run the first part of the command first if you want to spot check the output to confirm that it behaves as you expect (though remove the `includeAll` parameter if you don't want to spot check the entire list of operations)

By default, piped input will be processed serially (one after the other). If you want to speed things up you can use multiple workers (in this case let's use 5). We don't want to spam the platform, so we also set a delay of 250 milliseconds between each request processed by a single worker. Since we are potentially processing a large list of operations, we can also activate the progress indicator by using the `progress` parameter.

<CodeExample>

```bash
c8y operations list --status "PENDING" --dateTo "-14d" --includeAll |  # get list of operations
    c8y operations update \
        --status "FAILED" \
        --failureReason "Cancelled stale operation as it was older than 14 days" \
        --workers 5 \
        --delay 250 \
        --progress
```

```powershell
Get-OperationCollection -Status "PENDING" -DateTo "-14d" -IncludeAll |  # get list of operations
    Update-Operation `
        -Status "FAILED" `
        -FailureReason "Cancelled stale operation as it was older than 14 days" `
        -Workers 5 `
        -Delay 250 `
        -Progress
```

</CodeExample>

A few notes about the usage of the progress indicator. When using the `progress` parameter it will no longer write the response to standard output, however you can still redirect the output to file.

<CodeExample>

```bash
c8y operations list --status PENDING --dateFrom "-14d" |
    c8y operations update --progress > myoutput.json
```

```powershell
Get-OperationCollection -Status PENDING -DateFrom "-14d" | batch |
    Update-Operation -Status -Progress > myoutput.json
```

</CodeExample>

Before starting a command which will iterate over a large list of objects, it is helpful to first check how many objects exist to see if it aligns with your expectations to reduce the risk of modifying unexpected objects.

<CodeExample>

```bash
c8y operations list --status "PENDING" --dateTo "-14d" --withTotalPages --pageSize 1
```

```powershell
Get-OperationCollection -Status PENDING -DateFrom "-14d" -WithTotalPages -PageSize 1
```

</CodeExample>

## Piping to non-default flags (not supported in PSc8y)

:::note
Custom pipeline mapping is supported from `c8y` ≥ 2.17.0
:::

Most commands have a default flag which is marked as the default consumer of piped input. For most cases this is ok, however there are special cases where you might what to map the piped input to another flag.

The default pipeline variable can be changed by using the `-` (dash) character on the desired flag. The `-` value will be substituted with the piped input (standard input) line by line.

If the piped input is compressed JSON, then `c8y` will try to pick the correct property by a default list of properties defined internally on the command. If inbuilt properties do not match your use-case then you can specify a single or list of properties using the syntax `-.prop1,.prop2`, then `c8y` will take the value form the first matching property.

:::warning Not supported in PowerShell
Custom piped input to flag mapping is not supported in PSc8y (PowerShell) cmdlets as PowerShell does not support this.
:::

<CodeExample>

```bash
# Pick plain text input
echo "t12345" | c8y applications list --providedFor -
# => GET /application/applications?providedFor=t12345

# Pick providerFor value from json input from either from .name or .tenant properties
echo "{\"tenant\":\"t12345\"}" | c8y applications list --providedFor -.name,.tenant
# => GET /application/applications?providedFor=t12345
```

```powershell
# Not supported in PowerShell due to limitations in PowerShell itself
```

</CodeExample>

### Find any events leading up to an alarm

Image if there is a specific alarm and you would like to retrieve events leading up to the time of alarm. First you query for the alarm, and then pipe the value to the events command, and pipe the alarm's timestamp to the `dateFrom` flag.

<CodeExample>

```bash
c8y alarms list --type myCriticalAlarm --pageSize 1 \
| c8y events list --dateTo -.time

# If you want to limit to a single device,
# then you need to provide device in both commands
c8y alarms list --type myCriticalAlarm --pageSize 1 --device 12345 \
| c8y events list --dateTo -.time --device 12345
```

</CodeExample>

### Find devices by name

Map the piped input to the `--name` flag, which simplifies the command usage without having to explicitly pipe the whole inventory query (as `--query` is consumes the piped input by default).

<CodeExample>

```bash
echo -e "linuxdevice01\nlinuxdevice02" | c8y devices list --name -
# => GET /inventory/managedObjects?q=$filter=(name eq 'linuxdevice01') $orderby=name
# => GET /inventory/managedObjects?q=$filter=(name eq 'linuxdevice02') $orderby=name
```

</CodeExample>

### Assign multiple groups to the same user

`c8y userreferences addUserToGroup` normally accepts piped input via the `--user` flag, which results in add a single group to multiple users.

However if you would like to assign multiple groups to a single user, then the piped input can be assigned to the `--group` flag using the example below.

<CodeExample>

```bash
# Add assign multiple groups to a single user
echo -e "admins\nbusiness" | c8y userreferences addUserToGroup --user myuser@example.com --group -

# Or assign multiple users to a single group
echo -e "user01\nuser02" | c8y userreferences addUserToGroup --group admins
```

</CodeExample>


## Examples

### Piping ids

<CodeExample>

```bash
echo -e "12345\n22222" | c8y inventory get
```

```powershell
12345, 22222 | batch | Get-ManagedObject
```

</CodeExample>

### Piping names

<CodeExample>

```bash
echo -e "device01\ndevice02" | c8y inventory get
```

```powershell
"device01", "device02" | batch | Get-ManagedObject
```

</CodeExample>

### Creating devices

Use the `c8y util repeat` command to create 5 devices names, then pipe the names to `c8y devices create`

<CodeExample transform="false">

```bash
c8y util repeat 5 --format "device_%s%04s" |
    c8y devices create --template "{ other_props: {}, type: 'ci_Test' }" \
        --select id,name,type
```

```powershell
c8y util repeat 5 --format "device_%s%04s" |
    batch |
    New-Device -Template "{ other_props: {}, type: 'ci_Test' }" -Select id,name,type
```

</CodeExample>

```bash title="output"
| id                    | name                       | type                   |
|-----------------------|----------------------------|------------------------|
| 481033                | device_0001                | ci_Test                |
| 481034                | device_0002                | ci_Test                |
| 480956                | device_0003                | ci_Test                |
| 481035                | device_0004                | ci_Test                |
| 480860                | device_0005                | ci_Test                |
```

:::tip Shell Users
Alternatively the linux command `seq` can be used to create devices names to be piped to the create command.

```bash
seq -f device_%04g 6 10 |
    c8y devices create --template "{ other_props: {}, type: 'ci_Test' }" --select id,name,type
```

```bash title="output"
| id                    | name                       | type                   |
|-----------------------|----------------------------|------------------------|
| 481036                | device_0006                | ci_Test                |
| 480957                | device_0007                | ci_Test                |
| 481037                | device_0008                | ci_Test                |
| 480861                | device_0009                | ci_Test                |
| 480862                | device_0010                | ci_Test                |
```
:::

### Piping json

c8y also accepts piped json input (each line must represent a json object, this is automatically handled by c8y when piping to other c8y commands)

<CodeExample>

```bash
c8y inventory find --query="not(has(myTag))" |
    c8y inventory update --data "myTag={}" --dry
```

```powershell
Find-ManagedObject -Query="not(has(myTag))" | batch |
    Update-ManagedObject -Data "myTag={}" -Dry
```

</CodeExample>

Or if you are just want to update a single item

<CodeExample>

```bash
c8y inventory get --id=12345 | c8y inventory update --data "myTag={}" --dry
```

```powershell
Get-ManagedObject -Id 12345 | Update-ManagedObject -Data "myTag={}" -Dry
```

</CodeExample>

### Chaining commands and using templates

<CodeExample>

```bash
c8y devices create --name "myname" |
    c8y identity create --type c8y_Serial --template "{ externalId: input.value.name }"
```

```powershell
New-Device -Name "myname" |
    New-ExternalId -Type c8y_Serial -Template "{ externalId: input.value.name }"
```

</CodeExample>


```bash title="output"
| type            | externalId  | managedObject.id |
|-----------------|-------------|------------------|
| c8y_Serial      | myname      | 484203           |
```

### Piping inventory items and us multiple workers to speed up the action

<CodeExample>

```bash
c8y inventory list --pageSize 15 | c8y inventory get --workers 5
```

```powershell
Get-ManagedObjectCollection -PageSize 15 | batch | Get-ManagedObject -Workers 5
```

</CodeExample>

However if you are getting a list of devices, then you can just pipe the values directly

<CodeExample>

```bash
echo "1111\n2222" | c8y inventory get | c8y devices get --workers 1
```

```powershell
"1111", "2222" | batch | Get-ManagedObject -Workers 1
```

</CodeExample>

## Piping data from files

The following show examples of the support files inputs that can be piped to c8y commands. All of the examples retrieve operations related to each input device.

### List of ids

```json title="file: ids.txt (Comma Separated Variables (csv) - comma delimited)"
1111
2222
3333
4444    
```

<CodeExample>

```bash
cat input.txt | c8y operations list
```

```powershell
Get-Content input.txt | batch | Get-OperationCollection
```

</CodeExample>

### List of names

```json title="file: names.txt (Comma Separated Variables (csv) - comma delimited)"
MyActualName1
MyActualName2
MyActualName3
MyActualName4    
```

<CodeExample>

```bash
cat names.txt | c8y operations list
```

```powershell
Get-Content names.txt | batch | Get-OperationCollection
```

</CodeExample>

### CSV list (first column with id)

```csv title="file: input.txt (Comma Separated Variables (csv) - comma delimited)"
1111,MyActualName1
2222,MyActualName2
3333,MyActualName3
4444,MyActualName4    
```

<CodeExample>

```bash
cat input.txt | c8y operations list
```

```powershell
Get-Content input.txt | batch | Get-OperationCollection
```

</CodeExample>

### JSON Lines

```json title="file: input.json (JSON lines)"
{"name": "MyActualName1", "id": "1111", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_1"}}
{"name": "MyActualName2", "id": "2222", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_2"}}
{"name": "MyActualName3", "id": "3333", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_3"}}
{"name": "MyActualName4", "id": "4444", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_4"}}    
```

<CodeExample>

```bash
cat input.json | c8y operations list
```

```powershell
Get-Content input.json | batch | Get-OperationCollection
```

</CodeExample>
