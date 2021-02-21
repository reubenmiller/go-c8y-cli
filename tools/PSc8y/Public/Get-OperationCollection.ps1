# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-OperationCollection {
<#
.SYNOPSIS
Get a collection of operations based on filter parameters

.DESCRIPTION
Get a collection of operations based on filter parameters

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
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
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
        $BulkOperationId
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Agent")) {
            $Parameters["agent"] = PSc8y\Expand-Id $Agent
        }
        if ($PSBoundParameters.ContainsKey("DateFrom")) {
            $Parameters["dateFrom"] = $DateFrom
        }
        if ($PSBoundParameters.ContainsKey("DateTo")) {
            $Parameters["dateTo"] = $DateTo
        }
        if ($PSBoundParameters.ContainsKey("Status")) {
            $Parameters["status"] = $Status
        }
        if ($PSBoundParameters.ContainsKey("BulkOperationId")) {
            $Parameters["bulkOperationId"] = $BulkOperationId
        }

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
            | c8y operations list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | c8y operations list $c8yargs
        }
        
    }

    End {}
}
