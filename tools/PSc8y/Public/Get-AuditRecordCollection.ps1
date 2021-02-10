# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AuditRecordCollection {
<#
.SYNOPSIS
Get collection of (user) audits

.DESCRIPTION
Audit records contain information about modifications to other Cumulocity entities. For example the audit records contain each operation state transition, so they can be used to check when an operation transitioned from PENDING -> EXECUTING -> SUCCESSFUL.


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
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
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
        $Revert,

        # Maximum number of results
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateRange(1,2000)]
        [int]
        $PageSize,

        # Include total pages statistic
        [Parameter()]
        [switch]
        $WithTotalPages,

        # Get a specific page result
        [Parameter()]
        [int]
        $CurrentPage,

        # Maximum number of pages to retrieve when using -IncludeAll
        [Parameter()]
        [int]
        $TotalPages,

        # Include all results
        [Parameter()]
        [switch]
        $IncludeAll,

        # Show the full (raw) response from Cumulocity including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Write the response to file
        [Parameter()]
        [string]
        $OutputFile,

        # Ignore any proxy settings when running the cmdlet
        [Parameter()]
        [switch]
        $NoProxy,

        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Type")) {
            $Parameters["type"] = $Type
        }
        if ($PSBoundParameters.ContainsKey("User")) {
            $Parameters["user"] = $User
        }
        if ($PSBoundParameters.ContainsKey("Application")) {
            $Parameters["application"] = $Application
        }
        if ($PSBoundParameters.ContainsKey("DateFrom")) {
            $Parameters["dateFrom"] = $DateFrom
        }
        if ($PSBoundParameters.ContainsKey("DateTo")) {
            $Parameters["dateTo"] = $DateTo
        }
        if ($PSBoundParameters.ContainsKey("Revert")) {
            $Parameters["revert"] = $Revert
        }
        if ($PSBoundParameters.ContainsKey("PageSize")) {
            $Parameters["pageSize"] = $PageSize
        }
        if ($PSBoundParameters.ContainsKey("WithTotalPages") -and $WithTotalPages) {
            $Parameters["withTotalPages"] = $WithTotalPages
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        $Parameters["source"] = PSc8y\Expand-Id $Source

        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        Invoke-ClientCommand `
            -Noun "auditRecords" `
            -Verb "list" `
            -Parameters $Parameters `
            -Type "application/vnd.com.nsn.cumulocity.auditRecordCollection+json" `
            -ItemType "application/vnd.com.nsn.cumulocity.auditRecord+json" `
            -ResultProperty "auditRecords" `
            -Raw:$Raw `
            -CurrentPage:$CurrentPage `
            -TotalPages:$TotalPages `
            -IncludeAll:$IncludeAll
    }

    End {}
}
