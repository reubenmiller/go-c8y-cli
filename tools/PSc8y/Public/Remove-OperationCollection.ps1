# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-OperationCollection {
<#
.SYNOPSIS
Delete operation collection

.DESCRIPTION
Delete a collection of operations using a set of filter criteria. Be careful when deleting operations. Where possible update operations to FAILED (with a failure reason) instead of deleting them as it is easier to track.


.LINK
c8y operations deleteCollection

.EXAMPLE
PS> Remove-OperationCollection -Device "{{ randomdevice }}" -Status PENDING

Remove all pending operations for a given device


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
        $Status
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations deleteCollection"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y operations deleteCollection $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y operations deleteCollection $c8yargs
        }
        
    }

    End {}
}
