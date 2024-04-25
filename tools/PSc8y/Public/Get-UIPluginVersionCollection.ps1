# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UIPluginVersionCollection {
<#
.SYNOPSIS
Get plugin version collection

.DESCRIPTION
Get a collection of plugin versions by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_versions_list

.EXAMPLE
PS> Get-UIPluginVersionCollection -Plugin 1234

Get plugin versions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Plugin
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Plugin
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins versions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationVersionCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.applicationVersion+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions list $c8yargs
        }
        
    }

    End {}
}
