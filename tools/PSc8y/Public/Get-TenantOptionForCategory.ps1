# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantOptionForCategory {
<#
.SYNOPSIS
Get tenant options for category

.DESCRIPTION
Get tenant options for category

.EXAMPLE
PS> Get-TenantOptionForCategory -Category "c8y_cli_tests"

Get a list of options for a category


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant Option category (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Category
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantOptions getForCategory"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.optionCollection+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Category `
            | Group-ClientRequests `
            | c8y tenantOptions getForCategory $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Category `
            | Group-ClientRequests `
            | c8y tenantOptions getForCategory $c8yargs
        }
        
    }

    End {}
}
