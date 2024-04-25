# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-UIPluginVersion {
<#
.SYNOPSIS
Replace tags related to a plugin version

.DESCRIPTION
Replaces the tags of a given plugin version in your tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_versions_update

.EXAMPLE
PS> Update-UIPluginVersion -Plugin 1234 -Version 1.0 -Tags tag1,latest

Replace tags assigned to a version of a plugin


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

        # Version
        [Parameter()]
        [object[]]
        $Version,

        # Tag assigned to the version. Version tags must be unique across all versions and version fields of plugin versions
        [Parameter()]
        [string[]]
        $Tags
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins versions update"
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
            | c8y ui plugins versions update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions update $c8yargs
        }
        
    }

    End {}
}
