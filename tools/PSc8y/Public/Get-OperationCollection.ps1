# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-OperationCollection {
<#
.SYNOPSIS
Get operation collection

.DESCRIPTION
Get a collection of operations based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/operations_list

.EXAMPLE
PS> Get-OperationCollection -Status PENDING

Get a list of pending operations

.EXAMPLE
PS> Get-OperationCollection -Agent $Agent.id -Status PENDING

Get a list of pending operations for a given agent and all of its child devices

.EXAMPLE
PS> Get-OperationCollection -Device $Device.id -Status PENDING

Get a list of pending operations for a device

.EXAMPLE
PS> Get-DeviceCollection -Name $Agent2.name | Get-OperationCollection

Get operations from a device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Agent ID
        [Parameter()]
        [object[]]
        $Agent,

        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # The type of fragment that must be part of the operation. i.e. c8y_Restart
        [Parameter()]
        [string]
        $FragmentType,

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

        # Bulk operation id. Only retrieve operations related to the given bulk operation.
        [Parameter()]
        [string]
        $BulkOperationId,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.operationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.operation+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y operations list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y operations list $c8yargs
        }
        
    }

    End {}
}
