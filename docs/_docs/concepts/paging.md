---
layout: default
category: Concepts
title: Paging
---

### Limiting query results

#### Background

Cumulocity uses paging for any query which returns a collection of resources. For example, when getting a list of devices (managed objects), the query will limit the number of devices returns to the defined `pageSize`.

The `pageSize` has an upper limit of 2000 which is enforced on the server side.

In addition, Cumulocity also supports a `currentPage` parameter which can be used to control which page from the collection the server should return. The `currentPage` defaults 1, however can be changed to return a different page of the collection.

Please read the [Cumulocity IoT documentation](http://cumulocity.com/guides/reference/rest-implementation/#rest-usage) for futher details about paging.


#### Using paging on the command line

The c8y cli tool supports interfacting with Cumulocity's paging, and provides a few enhancements to help to get all of the data that you need to be productive.

The following command line arguments are supported by all collection related commands.

| PSc8y | c8y | Description |
|-------|---------|---------|
| -PageSize | `--pageSize` | Set the maximum number of results included in the response. Limited to 2000 by the server. |
| -CurrentPage | `--currentPage` | Page number (or slice) of the results which should be returned. Defaults to 1. |
| -TotalPages | `--totalPages` | Total number of pages to return. The pages will be collected serially starting from the `currentPage`  |
| -IncludeAll | `--includeAll` | Include all pages from the collection. IncludeAll will automatically set the pageSize to the maximum value `2000` regardless what value was given to `pageSize` |


A few examples will now be detailed to clarify the usage of the parameters in real life scenarios.

#### Example: Get all devices with a matching name prefix

If you have a large number of devices and you want to retrieve all of the results for some offline analysis.

The `includeAll` parameter is used to get all of the results.

**Powershell**

```powershell
Get-DeviceCollection -Name "MyDevices*" -IncludeAll
```

**Bash/zsh**

```sh
c8y devices list --name "MyDevices*" --includeAll
```

#### Iterating through the results

Instead of retrieving all of the devices at once, you can run a task on each paging result, and then move on to the next page, until all of the paging have been processed.

This has the advantages that all of the results do not need to be kept in memory.

**Example: Add a fragment to each device**

The following shows how ot add a fragment `myNewFragment` to each devices where the name starts with "My".

```powershell
Get-DeviceCollection -Name "My*" -IncludeAll |
    Update-ManagedObject -Data @{
            myNewFragment = @{
                # add timestamp (in ISO8601 format)
                fragmentCreationTime = Format-Date
            }
        }
```

**Notes**

* An `foreach` loop is need in the `ForEach-CumulocityPage` because the PowerShell iterator object `$_` will be an array of managed objects.


### Setting a default pageSize

The default pageSize can be controlled via the `settings.json` or in your session file. See [configuration options](https://reubenmiller.github.io/go-c8y-cli/docs/configuration/options/) for details.

### Notes / Warnings

* Retrieving all results will use more memory!!!
* When retrieving all results, it is recommended to use the maximum pages size rather than the default 5, otherwise you will have to send a lot of requests to get all of the results that you want
