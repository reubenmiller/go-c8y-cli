# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-AuditRecordCollection {
<#
.SYNOPSIS
Delete a collection of audit records

.DESCRIPTION
Important: This method has been deprecated and will be removed completely with the July 2020 release (10.6.6). With Cumulocity IoT >= 10.6.6 the deletion of audit logs will no longer be permitted. All DELETE requests to the audit API will return the error 405 Method not allowed. Note that retention rules still apply to audit logs and will delete audit log records older than the specified retention time.

.EXAMPLE
PS> Remove-AuditRecordCollection -Source $Device.id

Delete audit records from a device


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Source Id or object containing an .id property of the element that should be detected. i.e. AlarmID, or Operation ID. Note: Only one source can be provided
        [Parameter()]
        [string]
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
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Source")) {
            $Parameters["source"] = $Source
        }
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
        foreach ($item in @("")) {

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
                -Verb "deleteCollection" `
                -Parameters $Parameters `
                -Type "" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
