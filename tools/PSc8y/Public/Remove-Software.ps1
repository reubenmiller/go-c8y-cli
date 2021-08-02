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
        # Software Package (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Remove all versions recursively
        [Parameter()]
        [switch]
        $Cascade
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
