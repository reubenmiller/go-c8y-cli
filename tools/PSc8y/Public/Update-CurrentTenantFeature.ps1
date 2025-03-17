# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-CurrentTenantFeature {
<#
.SYNOPSIS
Update tenant features

.DESCRIPTION
Update tenant features

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_update

.EXAMPLE
PS> Update-CurrentTenantFeature -Key example -Active

Enable a feature in the current tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Feature ID / Key (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Key,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.tenant+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | Group-ClientRequests `
            | c8y features update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y features update $c8yargs
        }
        
    }

    End {}
}
