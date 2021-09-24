# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-FirmwareVersion {
<#
.SYNOPSIS
Delete firmware package version

.DESCRIPTION
Delete an existing firmware package version

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_versions_delete

.EXAMPLE
PS> Remove-FirmwareVersion -Id $mo.id

Delete a firmware version and all related versions

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-FirmwareVersion

Delete a firmware package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware Package version id or name (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Firmware package id or name (used to help completion be more accurate)
        [Parameter()]
        [object[]]
        $Firmware,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware versions delete"
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
            | c8y firmware versions delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y firmware versions delete $c8yargs
        }
        
    }

    End {}
}
