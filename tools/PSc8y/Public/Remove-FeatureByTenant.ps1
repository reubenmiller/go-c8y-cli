# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-FeatureByTenant {
<#
.SYNOPSIS
Removes the feature toggle override for a tenant of authenticated user

.DESCRIPTION
Removes the feature toggle override for a tenant of authenticated user

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_tenants_delete

.EXAMPLE
PS> Remove-FeatureByTenant -Id example -Tenant t12345

Remove the feature override from a given tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Feature ID/Key
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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features tenants delete"
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
            | c8y features tenants delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Tenant `
            | Group-ClientRequests `
            | c8y features tenants delete $c8yargs
        }
        
    }

    End {}
}
