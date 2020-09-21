# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-AuditRecord {
<#
.SYNOPSIS
Create a new audit record

.DESCRIPTION
Create a new audit record for a given action

.EXAMPLE
PS> New-AuditRecord -Type "ManagedObject" -Time "0s" -Text "Managed Object updated: my_Prop: value" -Source $Device.id -Activity "Managed Object updated" -Severity "information"

Create an audit record for a custom managed object update


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the type of this audit record. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Type,

        # Time of the audit record.
        [Parameter()]
        [string]
        $Time,

        # Text description of the audit record. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Text,

        # An optional ManagedObject that the audit record originated from (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Source,

        # The activity that was carried out. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Activity,

        # The severity of action: critical, major, minor, warning or information. (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('critical','major','minor','warning','information')]
        [string]
        $Severity,

        # The user responsible for the audited action.
        [Parameter()]
        [string]
        $User,

        # The application used to carry out the audited action.
        [Parameter()]
        [string]
        $Application,

        # Additional properties of the audit record.
        [Parameter()]
        [object]
        $Data,

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
        if ($PSBoundParameters.ContainsKey("Type")) {
            $Parameters["type"] = $Type
        }
        if ($PSBoundParameters.ContainsKey("Time")) {
            $Parameters["time"] = $Time
        }
        if ($PSBoundParameters.ContainsKey("Text")) {
            $Parameters["text"] = $Text
        }
        if ($PSBoundParameters.ContainsKey("Source")) {
            $Parameters["source"] = $Source
        }
        if ($PSBoundParameters.ContainsKey("Activity")) {
            $Parameters["activity"] = $Activity
        }
        if ($PSBoundParameters.ContainsKey("Severity")) {
            $Parameters["severity"] = $Severity
        }
        if ($PSBoundParameters.ContainsKey("User")) {
            $Parameters["user"] = $User
        }
        if ($PSBoundParameters.ContainsKey("Application")) {
            $Parameters["application"] = $Application
        }
        if ($PSBoundParameters.ContainsKey("Data")) {
            $Parameters["data"] = ConvertTo-JsonArgument $Data
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
                -Verb "create" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.auditRecord+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
