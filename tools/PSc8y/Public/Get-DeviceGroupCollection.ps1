# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceGroupCollection {
<#
.SYNOPSIS
Get device group collection

.DESCRIPTION
Get a collection of device groups based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_list

.EXAMPLE
PS> Get-DeviceGroupCollection -Name "MyGroup*"

Get a collection of device groups with names that start with "MyGroup"


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents,

        # Include names of child assets (only use where necessary as it is slow for large groups)
        [Parameter()]
        [switch]
        $WithChildren
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDeviceGroup+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devicegroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devicegroups list $c8yargs
        }
    }

    End {}
}
