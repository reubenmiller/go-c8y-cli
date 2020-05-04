# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AuditRecordCollection {
<#
.SYNOPSIS
Get collection of (user) audits

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

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
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
            -IncludeAll:$IncludeAll
    }

    End {}
}
