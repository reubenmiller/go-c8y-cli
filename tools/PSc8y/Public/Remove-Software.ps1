# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Software {
<#
.SYNOPSIS
Delete software package

.DESCRIPTION
Delete an existing software package

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_delete

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-Software -ForceCascade:$false

Delete a software package and all related versions

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-Software

Delete a software package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Software Package (managedObject) id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software delete"
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
            | c8y software delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y software delete $c8yargs
        }
        
    }

    End {}
}
