# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-FeatureByTenant {
<#
.SYNOPSIS
Set a feature override for a given tenant

.DESCRIPTION
Set a feature override for a given tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_tenants_update

.EXAMPLE
PS> Update-FeatureByTenant -Id example -Active -Tenant t12345

Enable a feature for a given specific tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Feature ID / Key (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Key,

        # Unique identifier of a Cumulocity tenant
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object]
        $Tenant,

        # Enable feature
        [Parameter()]
        [switch]
        $Active
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features tenants update"
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
            | c8y features tenants update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Tenant `
            | Group-ClientRequests `
            | c8y features tenants update $c8yargs
        }
        
    }

    End {}
}
