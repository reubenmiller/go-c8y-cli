# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-UIPluginVersion {
<#
.SYNOPSIS
Delete a specific version of a plugin

.DESCRIPTION
Delete a specific version of a plugin in your tenant, by a given tag or version

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_versions_delete

.EXAMPLE
PS> Remove-UIPluginVersion -Plugin 1234 -Tag tag1

Delete plugin version by tag

.EXAMPLE
PS> Remove-UIPluginVersion -Plugin 1234 -Version 1.0

Delete plugin version by version name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Plugin
        [Parameter()]
        [object[]]
        $Plugin,

        # Version, e.g. 1.0.0
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Version,

        # The tag of the plugin version
        [Parameter()]
        [string]
        $Tag
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins versions delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Version `
            | Group-ClientRequests `
            | c8y ui plugins versions delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Version `
            | Group-ClientRequests `
            | c8y ui plugins versions delete $c8yargs
        }
        
    }

    End {}
}
