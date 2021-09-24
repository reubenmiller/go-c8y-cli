# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Configuration {
<#
.SYNOPSIS
Delete configuration file

.DESCRIPTION
Delete an existing configuration file

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/configuration_delete

.EXAMPLE
PS> Remove-Configuration -Id $mo.id

Delete a configuration package (and any related binaries)

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-Configuration

Delete a configuration package (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $Device.id | Remove-Configuration -forceCascade:$false

Delete a configuration package but keep any related binaries


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Configuration file (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "configuration delete"
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
            | c8y configuration delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y configuration delete $c8yargs
        }
        
    }

    End {}
}
