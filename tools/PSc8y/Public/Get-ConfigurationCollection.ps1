# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ConfigurationCollection {
<#
.SYNOPSIS
Get configuration collection

.DESCRIPTION
Get a collection of configuration (managedObjects) based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/configuration_list

.EXAMPLE
PS> Get-ConfigurationCollection

Get a list of configuration files


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Configuration name filter
        [Parameter()]
        [string]
        $Name,

        # Configuration description filter
        [Parameter()]
        [string]
        $Description,

        # Configuration device type filter
        [Parameter()]
        [string]
        $DeviceType
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "configuration list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y configuration list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y configuration list $c8yargs
        }
    }

    End {}
}
