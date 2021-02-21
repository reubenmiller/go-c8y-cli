# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-TenantOptionBulk {
<#
.SYNOPSIS
Update multiple tenant options in provided category

.DESCRIPTION
Update multiple tenant options in provided category

.EXAMPLE
PS> Update-TenantOptionBulk -Category "c8y_cli_tests" -Data @{ $option5 = 0; $option6 = 1 }

Update multiple tenant options


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
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
        Get-ClientCommonParameters -Type "Update", "Template" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantOptions updateBulk"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Category)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Category `
            | c8y tenantOptions updateBulk $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Category `
            | c8y tenantOptions updateBulk $c8yargs
        }
        
    }

    End {}
}
