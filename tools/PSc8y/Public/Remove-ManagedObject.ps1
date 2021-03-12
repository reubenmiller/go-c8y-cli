# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ManagedObject {
<#
.SYNOPSIS
Delete managed object

.DESCRIPTION
Delete an existing managed object

.LINK
c8y inventory delete

.EXAMPLE
PS> Remove-ManagedObject -Id $mo.id

Delete a managed object

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-ManagedObject

Delete a managed object (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $Device.id | Remove-ManagedObject -Cascade

Delete a managed object and all child devices


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Remove all child devices and child assets will be deleted recursively. By default, the delete operation is propagated to the subgroups only if the deleted object is a group
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory delete"
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
            | c8y inventory delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y inventory delete $c8yargs
        }
        
    }

    End {}
}
