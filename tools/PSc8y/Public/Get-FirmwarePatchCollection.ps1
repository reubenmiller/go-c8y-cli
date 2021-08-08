# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-FirmwarePatchCollection {
<#
.SYNOPSIS
Get firmware patch collection

.DESCRIPTION
Get a collection of firmware patches (managedObjects) based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_patches_list

.EXAMPLE
PS> Get-FirmwarePatchCollection -FirmwareId 12345

Get a list of firmware patches related to a firmware package

.EXAMPLE
PS> Get-FirmwarePatchCollection -FirmwareId 12345 -Dependency '1.*'

Get a list of firmware patches where the dependency version starts with "1."


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

        # Patch dependency version
        [Parameter()]
        [string]
        $Dependency,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware patches list"
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
            | c8y firmware patches list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $FirmwareId `
            | Group-ClientRequests `
            | c8y firmware patches list $c8yargs
        }
        
    }

    End {}
}
