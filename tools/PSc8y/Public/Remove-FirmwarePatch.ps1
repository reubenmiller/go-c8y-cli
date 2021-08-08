# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-FirmwarePatch {
<#
.SYNOPSIS
Delete firmware package version patch

.DESCRIPTION
Delete an existing firmware package version patch

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_patches_delete

.EXAMPLE
PS> Remove-FirmwarePatch -Id $mo.id

Delete a firmware package version patch

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-FirmwarePatch

Delete a firmware patch (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $Device.id | Remove-FirmwarePatch -ForceCascade

Delete a firmware patch and related binary


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware patch (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Firmware id (used to help completion be more accurate)
        [Parameter()]
        [object[]]
        $FirmwareId,

        # Remove version and any related binaries
        [Parameter()]
        [switch]
        $ForceCascade
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware patches delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y firmware patches delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y firmware patches delete $c8yargs
        }
        
    }

    End {}
}
