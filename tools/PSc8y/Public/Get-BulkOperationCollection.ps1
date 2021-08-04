# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-BulkOperationCollection {
<#
.SYNOPSIS
Get bulk operation collection

.DESCRIPTION
Get a collection of bulk operations

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/bulkoperations_list

.EXAMPLE
PS> Get-BulkOperationCollection

Get a list of bulk operations


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Include CANCELLED bulk operations
        [Parameter()]
        [switch]
        $WithDeleted,

        # Start date or date and time of the bulk operation
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of the bulk operation
        [Parameter()]
        [string]
        $DateTo,

        # Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.
        [Parameter()]
        [ValidateSet('CANCELED','SCHEDULED','EXECUTING','EXECUTING_WITH_ERROR','FAILED')]
        [string[]]
        $Status
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkoperations list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.bulkOperationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.bulkoperation+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y bulkoperations list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y bulkoperations list $c8yargs
        }
    }

    End {}
}
