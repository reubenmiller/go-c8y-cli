# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Tenant {
<#
.SYNOPSIS
Update tenant

.EXAMPLE
PS> Update-Tenant -Id mycompany -ContactName "John Smith"

Update a tenant by name (from the mangement tenant)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Id,

        # Company name. Maximum 256 characters
        [Parameter()]
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

        # A set of custom properties of the tenant
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
        $Parameters["id"] = PSc8y\Expand-Id $Id

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
            -Verb "update" `
            -Parameters $Parameters `
            -Type "application/vnd.com.nsn.cumulocity.tenant+json" `
            -ItemType "" `
            -ResultProperty "" `
            -Raw:$Raw `
            -IncludeAll:$IncludeAll
    }

    End {}
}
