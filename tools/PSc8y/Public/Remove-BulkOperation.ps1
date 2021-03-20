# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-BulkOperation {
<#
.SYNOPSIS
Delete bulk operation

.DESCRIPTION
Delete bulk operation/s. Only bulk operations that are in ACTIVE or IN_PROGRESS can be deleted

.LINK
c8y bulkoperations delete

.EXAMPLE
PS> Remove-BulkOperation -Id $BulkOp.id

Remove bulk operation by id


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Bulk Operation id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkoperations delete"
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
            | c8y bulkoperations delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y bulkoperations delete $c8yargs
        }
        
    }

    End {}
}
