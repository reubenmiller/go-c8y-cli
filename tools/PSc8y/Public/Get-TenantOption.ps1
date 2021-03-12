# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantOption {
<#
.SYNOPSIS
Get tenant option

.DESCRIPTION
Get an existing tenant option

.LINK
c8y tenantOptions get

.EXAMPLE
PS> Get-TenantOption -Category "c8y_cli_tests" -Key "$option2"

Get a tenant option


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant Option category (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Category,

        # Tenant Option key (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Key
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantOptions get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | Group-ClientRequests `
            | c8y tenantOptions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y tenantOptions get $c8yargs
        }
        
    }

    End {}
}
