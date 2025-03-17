# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-CurrentTenantFeature {
<#
.SYNOPSIS
Removes the feature toggle override for a tenant of authenticated user

.DESCRIPTION
Removes the feature toggle override for a tenant of authenticated user

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/features_delete

.EXAMPLE
PS> Remove-CurrentTenantFeature -Key example

Remove the feature override for the current tenant


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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "features delete"
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
            | c8y features delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y features delete $c8yargs
        }
        
    }

    End {}
}
