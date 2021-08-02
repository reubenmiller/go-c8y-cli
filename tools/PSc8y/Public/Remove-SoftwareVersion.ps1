# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-SoftwareVersion {
<#
.SYNOPSIS
Delete software package version

.DESCRIPTION
Delete an existing software package version

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_versions_delete

.EXAMPLE
PS> Remove-Software -Id $mo.id

Delete a software package

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-Software

Delete a software package (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $Device.id | Remove-Software -Cascade

Delete a software package and all related versions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Software Package version (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Software package id (used to help completion be more accurate)
        [Parameter()]
        [object[]]
        $SoftwareId,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software versions delete"
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
            | c8y software versions delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y software versions delete $c8yargs
        }
        
    }

    End {}
}
