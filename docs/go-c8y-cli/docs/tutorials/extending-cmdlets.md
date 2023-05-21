---
layout: powershell
category: Tutorials - Powershell
title: Extending cmdlets using scripts and Functions
---

:::tip
It is recommended to use [extensions](/docs/concepts/extensions/) instead of creating your own PowerShell module. Extensions are available in go-c8y-cli >= 2.30.0 and provide a more portable/native experience. They are just as easy to install and share, and can be used in any shell.

Or just check out the [Tutorials](/docs/tutorials/extensions/).
:::

## Example 1: Script to export data for specific devices

### Goal

This example exports data from Cumulocity into a format that can be used for offline analysis. THe script will show how to export the following device information, however it can be extended to included any properties that you want.

* device id
* device name
* creation time
* owner
* availability status

Powershell contains useful cmdlet which handle importing and exporting command formats such as CSV and JSON. In this example the `Export-Csv` cmdlet will be used to convert data to csv.

### Useful cmdlets

The following cmdlets will be used in this example.

**PSc8y cmdlets**

* `Get-DeviceCollection` - Get a collection of devices using filter criteria

**Native cmdlets**

* `Get-Date` - Get current date / timestamp
* `Export-Csv` - Export data to CSV format
* `Select-Object` - Select a subset of properties from an object to only focus on data for the current task

### Script

Below shows the full script, following with the usage:

```powershell
[cmdletbinding()]
Param(
    # Device type. Only included devices with the specified type
    [Parameter(
        Mandatory = $true
    )]
    [string] $DeviceType,

    # Report Name.
    # The timestamp will be appended to the report name along with
    # the ".csv" file extension
    [string] $ReportName = "MyCustomReport"
)

# Step 1: Filter devices by type
[array] $Devices = Get-DeviceCollection -Type $DeviceType -PageSize 2000

# Create unique report name
$ReportTimeStamp = Get-Date -Format "yyy-MM-d_Hmmss"
$ReportName = "${ReportName}_${ReportTimeStamp}.csv"

# Step 2: Export selected data
$Devices |
    Select-Object id, name, type, owner, creationTime |
    Export-Csv -Path $ReportName

#
# Or if you want to include nested properties you can provide a
# complex property definition to Select-Object

# Alternative Step 2: Nested property to access .c8y_Availability.status
$status = @{
    l = "status"
    e = { $_.c8y_Availability.status }
}

$Devices |
    Select-Object -property id, name, owner, creationTime, $status |
    Export-Csv -Path $ReportName

```

### Usage

The script can then be called from the command line:

```powershell
./Export-CustomReport.ps1 -DeviceType "exampleType"
```

The script would create a CSV file containing the exported report information similar to the following:

*MyCustomReport_2020-09-20_00547.csv*

```csv
"id","name","type","owner","creationTime"
"2741586","example-01","exampleType","device_example-01","18.09.2020 18:24:12"
```

---

## Example 2: Cancel operations which are stuck in EXECUTING

The goal of this script will be to cancel operations assigned to a device that are in the `EXECUTING` status.

Cancelling the operation actually just means that the operation should be set to the status `FAILED`.

Like any task it is helpful to break it down into smaller parts.
1. Get the operation which are in EXECUTING
2. Set the operations returned from step 1 to FAILED

#### Used cmdlets

* `Expand-Device` - Normalize processing of devices by allowing accessing by id, name or objects
* `Get-OperationCollection` - Get a list of operations 
* `Update-Operation` - Change the operation status to `FAILED`


### Script (Advanced Function)

**Invoke-CancelOperations.ps1**

```powershell
Function Invoke-CancelOperations {
<# 
.SYNOPSIS
Cancel operations which are currently in EXECUTING status
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Device
        [Parameter(
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Mandatory = $true
        )]
        [object[]] $Device,

        # Don't prompt for confirmation
        [switch] $Force
    )

    Process {
        foreach ($iDevice in (Expand-Device $Device)) {

            [array] $operations = Get-OperationCollection `
                -Device $iDevice `
                -Status EXECUTING
            
            if ($operations.Count -gt 0) {
                $operations | Update-Operation `
                    -Status FAILED `
                    -FailureReason "Manually cancelled" `
                    -Force:$Force `
                    -Verbose:$VerbosePreference `
                    -Dry:$WhatIfPreference
            } else {
                Write-Warning ("Device [{0} ({1})] does not have any operations in [$Status] status" -f @(
                    $iDevice.name,
                    $iDevice.id
                ))
            }
        }
    }
}
```

### Usage

Because the above example is written as an advanced function, it needs to be imported before it can be used. The file can be imported using the command `. ./Invoke-CancelOperations.ps1`. It is pure convention that the function name is the same as the file name. A `.ps1` file can contain many advanced functions, however it is recommended to only store one per file. This layout will make it easier to convert them to a PowerShell module in the future.

```powershell
# Import (dot sourcing) your script
. ./Invoke-CancelOperations.ps1

# Use the function
Invoke-CancelOperations -Device 12345

# Use function with piped arguments
12345, 232323 | Invoke-CancelOperations

# Use function with piped devices
devices *example* | Invoke-CancelOperations
```

---

## Managing your scripts

It is common practice to automate the importing of commonly used scripts into your PowerShell profile. This means that everytime you open up a PowerShell concole, then all of your scripts will be imported and ready to use.

In order to make the importing of custom powershell extensible, you can add some code to you profile which will automatically import all `.ps1` files contained in a pre-defined folder. Note: `.ps1` is extension name for Powershell scripts.

1. Create a `script` folder in your home folder. Check the output of `$HOME` in PowerShell to show the path.

2. Get the path to your PowerShell profile. The file path can be shown by executing `$PROFILE`

3. Add the following code to your PowerShell profile

    ```powershell
    $ScriptSourceFolder = "$HOME/scripts"

    # Look for all .ps1 files, and import them
    Get-ChildItem -Path $ScriptSourceFolder -Filter "*.ps1" -Recurse |
        Foreach-Object {
            Write-Host ("importing [{0}]" -f $_.FullName) -ForegroundColor DarkCyan
            . $_.FullName
        }
    ```

4. Start a new Powershell profile. You should see some console messages showing which files were imported. If you don't see anything, then you probably have not added any `.ps1` scripts in the correct folder.
