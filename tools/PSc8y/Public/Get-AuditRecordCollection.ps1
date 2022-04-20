﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AuditRecordCollection {
<#
.SYNOPSIS
Get audit record collection

.DESCRIPTION
Audit records contain information about modifications to other Cumulocity entities.

For example the audit records contain each operation state transition, so they can be used to check when an operation transitioned from PENDING -> EXECUTING -> SUCCESSFUL.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/auditrecords_list

.EXAMPLE
PS> Get-AuditRecordCollection -PageSize 100

Get a list of audit records

.EXAMPLE
PS> Get-AuditRecordCollection -Source $Device2.id

Get a list of audit records related to a managed object

.EXAMPLE
PS> Get-Operation -Id $Operation.id | Get-AuditRecordCollection

Get a list of audit records related to an operation


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Source Id or object containing an .id property of the element that should be detected. i.e. AlarmID, or Operation ID. Note: Only one source can be provided
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Source,

        # Type
        [Parameter()]
        [string]
        $Type,

        # Username
        [Parameter()]
        [string]
        $User,

        # Application
        [Parameter()]
        [string]
        $Application,

        # Start date or date and time of audit record occurrence.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of audit record occurrence.
        [Parameter()]
        [string]
        $DateTo,

        # Return the newest instead of the oldest audit records. Must be used with dateFrom and dateTo parameters
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "auditrecords list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.auditRecordCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.auditRecord+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Source `
            | Group-ClientRequests `
            | c8y auditrecords list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Source `
            | Group-ClientRequests `
            | c8y auditrecords list $c8yargs
        }
        
    }

    End {}
}
