# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-TenantOptionEditable {
<#
.SYNOPSIS
Update tenant option editibility

.DESCRIPTION
Required role:: ROLE_OPTION_MANAGEMENT_ADMIN, Required tenant management Example Request:: Update access.control.allow.origin option.


.EXAMPLE
PS> Update-TenantOptionEditable -Category "c8y_cli_tests" -Key "$option8" -Editable "true"

Update editable property for an existing tenant option


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
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

        # Whether the tenant option should be editable or not (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('true','false')]
        [string]
        $Editable
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Category")) {
            $Parameters["category"] = $Category
        }
        if ($PSBoundParameters.ContainsKey("Editable")) {
            $Parameters["editable"] = $Editable
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantOptions updateEdit"
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
            $items = (PSc8y\Expand-Id $Key)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | c8y tenantOptions updateEdit $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | c8y tenantOptions updateEdit $c8yargs
        }
        
    }

    End {}
}
