# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Tenant {
<#
.SYNOPSIS
Update tenant

.DESCRIPTION
Update an existing tenant

.LINK
c8y tenants update

.EXAMPLE
PS> Update-Tenant -Id mycompany -ContactName "John Smith"

Update a tenant by name (from the management tenant)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
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
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

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

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y tenants update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y tenants update $c8yargs
        }
        
    }

    End {}
}
