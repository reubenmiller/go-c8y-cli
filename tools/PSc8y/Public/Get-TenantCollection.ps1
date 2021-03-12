# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantCollection {
<#
.SYNOPSIS
Get tenant collection

.DESCRIPTION
Get collection of tenants

.LINK
c8y tenants list

.EXAMPLE
PS> Get-TenantCollection

Get a list of tenants


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.tenantCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.tenant+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenants list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenants list $c8yargs
        }
    }

    End {}
}
