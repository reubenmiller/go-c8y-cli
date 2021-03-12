# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-TenantOption {
<#
.SYNOPSIS
Update tenant option

.DESCRIPTION
Update an existing tenant option

.LINK
c8y tenantOptions update

.EXAMPLE
PS> Update-TenantOption -Category "c8y_cli_tests" -Key "$option4" -Value "0"

Update a tenant option


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
        $Key,

        # New value (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Value
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantOptions update"
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
            | c8y tenantOptions update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y tenantOptions update $c8yargs
        }
        
    }

    End {}
}
