# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-FirmwarePatch {
<#
.SYNOPSIS
Get firmware patch

.DESCRIPTION
Get an existing firmware patch

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_patches_get

.EXAMPLE
PS> Get-FirmwarePatch -Id $mo.id

Get a firmware patch

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Get-FirmwarePatch

Get a firmware package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware Package version (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Firmware package id (used to help completion be more accurate)
        [Parameter()]
        [object[]]
        $FirmwareId,

        # Include parent references
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware patches get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObject+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y firmware patches get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y firmware patches get $c8yargs
        }
        
    }

    End {}
}
