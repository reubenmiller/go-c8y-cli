# Code generated from specification version 1.0.0: DO NOT EDIT
Function Invoke-CancelOperation {
<#
.SYNOPSIS
Cancel operation

.DESCRIPTION
Cancel an operation. This is a convenience command to set an operation to the FAILED status along with a sensible default failure reason.
Note: Cancelling an operation does not guarantee that any client that is already processing the operation will stop and in
normal circumstances this command should be used sparingly.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/operations_cancel

.EXAMPLE
PS> Invoke-CancelOperation -Id {{ NewOperation }}

Cancel an operation

.EXAMPLE
PS> Get-OperationCollection -Device $Agent.id -Status PENDING | Invoke-CancelOperation -FailureReason "manually cancelled"

Cancel multiple operations


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Operation id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Reason for the failure
        [Parameter()]
        [string]
        $FailureReason
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations cancel"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.operation+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y operations cancel $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y operations cancel $c8yargs
        }
        
    }

    End {}
}
