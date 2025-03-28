# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantTFASetting {
<#
.SYNOPSIS
Get Tenant TFA setting

.DESCRIPTION
Get a Tenant's Two-Factor-Authentication setting

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenants_tfa_get

.EXAMPLE
PS> Get-TenantTFASetting

Get the Two-Factor-Authentication setting of a tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant id
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenants tfa get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenants tfa get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenants tfa get $c8yargs
        }
    }

    End {}
}
