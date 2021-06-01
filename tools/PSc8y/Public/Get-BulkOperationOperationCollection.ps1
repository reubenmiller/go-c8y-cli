# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-BulkOperationOperationCollection {
<#
.SYNOPSIS
Get operations collection

.DESCRIPTION
Get a collection of operations related to a bulk operation

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/bulkoperations_listOperations

.EXAMPLE
PS> Get-BulkOperationOperationCollection -Id 10 -Status PENDING

Get a list of pending operations from bulk operation with id 10

.EXAMPLE
PS> Get-BulkOperationCollection | Where-Object { $_.status -eq "IN_PROGRESS" } | Get-BulkOperationOperationCollection -Status PENDING

Get all pending operations from all bulk operations which are still in progress (using pipeline)

.EXAMPLE
PS> Get-BulkOperationCollection | Get-BulkOperationOperationCollection -status EXECUTING --dateTo "-10d" | Update-Operation -Status FAILED -FailureReason "Manually cancelled stale operation"

Check all bulk operations if they have any related operations still in executing state and were created more than 10 days ago, then cancel it with a custom message


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Bulk operation id. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Start date or date and time of operation.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of operation.
        [Parameter()]
        [string]
        $DateTo,

        # Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.
        [Parameter()]
        [ValidateSet('PENDING','EXECUTING','SUCCESSFUL','FAILED')]
        [string]
        $Status,

        # Sort operations newest to oldest. Must be used with dateFrom and/or dateTo parameters
        [Parameter()]
        [switch]
        $Revert
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkoperations listOperations"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.operationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.operation+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y bulkoperations listOperations $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y bulkoperations listOperations $c8yargs
        }
        
    }

    End {}
}
