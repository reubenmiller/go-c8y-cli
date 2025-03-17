# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-FeatureByTenant {
<#
.SYNOPSIS
Get features toggle by tenant

.DESCRIPTION
Retrieve a list of all existing feature toggle value overrides for all tenants

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_tenants_list

.EXAMPLE
PS> Get-FeatureByTenant -Id example

Get the feature toggle override status for each child tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Feature ID/Key
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Key
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features tenants list"
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
            | c8y features tenants list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y features tenants list $c8yargs
        }
        
    }

    End {}
}
