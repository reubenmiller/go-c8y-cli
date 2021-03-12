# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-TenantOption {
<#
.SYNOPSIS
Delete tenant option

.DESCRIPTION
Delete an existing tenant option

.LINK
c8y tenantOptions delete

.EXAMPLE
PS> Remove-TenantOption -Category "c8y_cli_tests" -Key "$option3"

Delete a tenant option


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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantOptions delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | Group-ClientRequests `
            | c8y tenantOptions delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y tenantOptions delete $c8yargs
        }
        
    }

    End {}
}
