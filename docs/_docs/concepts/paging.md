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

```powershell
Get-DeviceCollection -Name "My*" -IncludeAll | ForEach-CumulocityPage {
    foreach ($item in $_) {
        Write-Host ("Adding fragment to device: {0} ({1})" -f $item.name, $item.id)

        Update-ManagedObject -Id $item.id -Data @{
            myNewFragment = @{
                # add timestamp (in ISO8601 format)
                fragmentCreationTime = Format-Date
            }
        }
    }
}
```

**Notes**

* An `foreach` loop is need in the `ForEach-CumulocityPage` because the PowerShell iterator object `$_` will be an array of managed objects.

### Alternative interface???

I'm not 100% sure if this is actually possible, if so it would definitely only be a powershell thing.

The idea would be using pwsh pipelines, the execution order would be something like:

1. Get the first page, pass it to Update-ManagedObject immediately
2. `Expand-CumulocityPage` would get the next page, then pass it to Update-ManagedObject

So you can do it in batches...but again, need to check if this is actually possible.

```powershell
Get-DeviceCollection -Name "My*" -PageSize 2000 |
    Expand-CumulocityPage |
    Update-ManagedObject @{
        myNewFragment = @{
            # add timestamp (in ISO8601 format)
            fragmentCreationTime = Format-Date
        }
    }
```

### Setting a default pageSize

```sh
$PSDefaultParameterValues = @{"*:PageSize" = 2000};
```

#### Notes / Warnings

* Retrieving all results will use more memory!!!
* When retrieving all results, it is recommended to use the maximum pages size rather than the default 5, otherwise you will have to send a lot of requests to get all of the results that you want
