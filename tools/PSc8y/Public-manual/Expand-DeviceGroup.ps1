Function Expand-DeviceGroup {
<#
.SYNOPSIS
Expand a list of device groups

.DESCRIPTION
Expand a list of device groups replacing any ids or names with the actual user object.

.NOTES
If the given object is already an user object, then it is added with no additional lookup

.EXAMPLE
Expand-DeviceGroup "myGroup"

Retrieve the user objects by name or id

.EXAMPLE
Get-DeviceGroupCollection *test* | Expand-DeviceGroup

Get all the device groups (with "test" in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

#>
    [cmdletbinding()]
    Param(
        # List of ids, names or user objects
        [Parameter(
            Mandatory = $true,
            ValueFromPipeline = $true,
            Position = 0
        )]
        [object[]] $InputObject
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        [array] $AllItems = foreach ($item in $InputObject) {
            if (($item -is [string])) {
                if ($item -match "\*") {
                    PSc8y\Get-DeviceGroupCollection -Name $item -WhatIf:$false -PageSize 100
                }
                else {
                    $item
                }
            }
            else {
                $item
            }
        }

        $AllItems
    }
}
