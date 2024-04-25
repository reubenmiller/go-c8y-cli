# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UIPluginVersion {
<#
.SYNOPSIS
Get a specific version of a plugin

.DESCRIPTION
Retrieve the selected version of a plugin in your tenant. To select the version, use only the version or only the tag query parameter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_versions_get

.EXAMPLE
PS> Get-UIPluginVersion -Plugin 1234 -Tag tag1

Get plugin version by tag

.EXAMPLE
PS> Get-UIPluginVersion -Plugin 1234 -Version 1.0

Get plugin version by version name


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
        $Plugin,

        # The version field of the plugin version
        [Parameter()]
        [object[]]
        $Version,

        # The tag of the plugin version
        [Parameter()]
        [string]
        $Tag
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins versions get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationVersion+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions get $c8yargs
        }
        
    }

    End {}
}
