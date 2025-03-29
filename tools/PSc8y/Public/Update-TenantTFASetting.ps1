# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-TenantTFASetting {
<#
.SYNOPSIS
Update Tenant TFA setting

.DESCRIPTION
Update a Tenant's Two-Factor-Authentication setting

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenants_tfa_update

.EXAMPLE
PS> Update-TenantTFASetting -Strategy TOTP

Update the Tenant's TFA setting of the current tenant

.EXAMPLE
PS> Update-TenantTFASetting -Tenant t12345 -Strategy TOTP

Update the Tenant's TFA setting to use time based one-time-password


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id. Defaults to current tenant
        [Parameter()]
        [object]
        $Tenant,

        # Two-factor authentication strategy
        [Parameter()]
        [ValidateSet('SMS','TOTP')]
        [string]
        $Strategy
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants tfa update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenants tfa update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenants tfa update $c8yargs
        }
    }

    End {}
}
