﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Operation {
<#
.SYNOPSIS
Update operation

.DESCRIPTION
Update an operation. This is commonly used to change an operation's status. For example the operation can be set to FAILED along with a failure reason.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/operations_update

.EXAMPLE
PS> Update-Operation -Id {{ NewOperation }} -Status EXECUTING

Update an operation

.EXAMPLE
PS> Get-OperationCollection -Device $Agent.id -Status PENDING | Update-Operation -Status FAILED -FailureReason "manually cancelled"

Update multiple operations


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

        # Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.
        [Parameter()]
        [ValidateSet('PENDING','EXECUTING','SUCCESSFUL','FAILED')]
        [string]
        $Status,

        # Reason for the failure. Use when setting status to FAILED
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "operations update"
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
            | c8y operations update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y operations update $c8yargs
        }
        
    }

    End {}
}
