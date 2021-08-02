# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-FirmwareVersionCollection {
<#
.SYNOPSIS
Get firmware package version collection

.DESCRIPTION
Get a collection of firmware package versions (managedObjects) based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_versions_list

.EXAMPLE
PS> Get-FirmwareVersionCollection

Get a list of firmware package versions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware package id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $FirmwareId,

        # Include parent references
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware versions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $FirmwareId `
            | Group-ClientRequests `
            | c8y firmware versions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $FirmwareId `
            | Group-ClientRequests `
            | c8y firmware versions list $c8yargs
        }
        
    }

    End {}
}
