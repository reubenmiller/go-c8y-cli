# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Tenant {
<#
.SYNOPSIS
New tenant

.EXAMPLE
PS> New-Tenant -Company "mycompany" -Domain "mycompany" -AdminName "admin" -Password "mys3curep9d8"
Create a new tenant (from the management tenant)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Company name. Maximum 256 characters (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Company,

        # Domain name to be used for the tenant. Maximum 256 characters (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Domain,

        # Username of the tenant administrator
        [Parameter()]
        [string]
        $AdminName,

        # Password of the tenant administrator
        [Parameter()]
        [string]
        $AdminPass,

        # A contact name, for example an administrator, of the tenant
        [Parameter()]
        [string]
        $ContactName,

        # An international contact phone number
        [Parameter()]
        [string]
        $ContactPhone,

        # The tenant ID. This should be left bank unless you know what you are doing. Will be auto-generated if not present.
        [Parameter()]
        [string]
        $TenantId,

        # A set of custom properties of the tenant
        [Parameter()]
        [object]
        $Data,

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
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Company")) {
            $Parameters["company"] = $Company
        }
        if ($PSBoundParameters.ContainsKey("Domain")) {
            $Parameters["domain"] = $Domain
        }
        if ($PSBoundParameters.ContainsKey("AdminName")) {
            $Parameters["adminName"] = $AdminName
        }
        if ($PSBoundParameters.ContainsKey("AdminPass")) {
            $Parameters["adminPass"] = $AdminPass
        }
        if ($PSBoundParameters.ContainsKey("ContactName")) {
            $Parameters["contactName"] = $ContactName
        }
        if ($PSBoundParameters.ContainsKey("ContactPhone")) {
            $Parameters["contactPhone"] = $ContactPhone
        }
        if ($PSBoundParameters.ContainsKey("TenantId")) {
            $Parameters["tenantId"] = $TenantId
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
                -Noun "tenants" `
                -Verb "create" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.tenant+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
