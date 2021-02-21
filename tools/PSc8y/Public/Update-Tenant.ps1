# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Tenant {
<#
.SYNOPSIS
Update tenant

.DESCRIPTION
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

        # Domain name to be used for the tenant. Maximum 256 characters
        [Parameter()]
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
        $ContactPhone
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template" -BoundParameters $PSBoundParameters
    }

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

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.tenant+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Id)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | c8y tenants update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | c8y tenants update $c8yargs
        }
        
    }

    End {}
}
