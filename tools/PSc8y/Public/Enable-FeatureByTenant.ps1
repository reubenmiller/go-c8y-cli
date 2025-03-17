# Code generated from specification version 1.0.0: DO NOT EDIT
Function Enable-FeatureByTenant {
<#
.SYNOPSIS
Enable a feature override for a given tenant

.DESCRIPTION
Enable a feature override for a given tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_tenants_enable

.EXAMPLE
PS> Enable-FeatureByTenant -Id example -Active -Tenant t12345

Enable a feature for a given tenant


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features tenants enable"
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
            | c8y features tenants enable $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Tenant `
            | Group-ClientRequests `
            | c8y features tenants enable $c8yargs
        }
        
    }

    End {}
}
