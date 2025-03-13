# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentTenantFeatureCollection {
<#
.SYNOPSIS
Get feature list for current tenant

.DESCRIPTION
Get feature list for current tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_list

.EXAMPLE
PS> Get-CurrentTenantFeatureCollection

Get a list of features for the current tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y features list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y features list $c8yargs
        }
    }

    End {}
}
