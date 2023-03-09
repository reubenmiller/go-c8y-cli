# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-DeviceSoftware {
<#
.SYNOPSIS
Add software package

.DESCRIPTION
Add software packages to a device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_software_add

.EXAMPLE
PS> Add-DeviceSoftware -Device 12345 -Name myapp -Version 1.0.2

Add software to a device


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device
        [Parameter()]
        [object[]]
        $Device,

        # Software name
        [Parameter()]
        [string]
        $Name,

        # Software version
        [Parameter()]
        [string]
        $Version,

        # Software url
        [Parameter()]
        [string]
        $Url,

        # Software type, e.g. apt
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices software add"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y devices software add $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y devices software add $c8yargs
        }
    }

    End {}
}
