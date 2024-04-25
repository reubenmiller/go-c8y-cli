# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-UIExtensionVersion {
<#
.SYNOPSIS
Replace tags related to an extension version

.DESCRIPTION
Replaces the tags of a given extension version in your tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_versions_update

.EXAMPLE
PS> Update-UIExtensionVersion -Extension 1234 -Version 1.0 -Tags tag1,latest

Replace tags assigned to a version of an extension


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Extension
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Extension,

        # Version
        [Parameter()]
        [object[]]
        $Version,

        # Tag assigned to the version. Version tags must be unique across all versions and version fields of extension versions
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
            $Extension `
            | Group-ClientRequests `
            | c8y ui plugins versions update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Extension `
            | Group-ClientRequests `
            | c8y ui plugins versions update $c8yargs
        }
        
    }

    End {}
}
