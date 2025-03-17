# Code generated from specification version 1.0.0: DO NOT EDIT
Function Disable-FeatureByTenant {
<#
.SYNOPSIS
Disable a feature override for a given tenant

.DESCRIPTION
Disable a feature override for a given tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_tenants_disable

.EXAMPLE
PS> Disable-FeatureByTenant -Id example -Active -Tenant t12345

Disable a feature in the current tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Feature ID / Key
        [Parameter()]
        [string]
        $Key,

        # Unique identifier of a Cumulocity tenant
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features tenants disable"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Tenant `
            | Group-ClientRequests `
            | c8y features tenants disable $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Tenant `
            | Group-ClientRequests `
            | c8y features tenants disable $c8yargs
        }
        
    }

    End {}
}
