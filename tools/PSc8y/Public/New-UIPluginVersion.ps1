# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-UIPluginVersion {
<#
.SYNOPSIS
Create a new version of a plugin

.DESCRIPTION
Uploaded version and tags can only contain upper and lower case letters, integers and ., +, -. Other characters are prohibited.

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_versions_create

.EXAMPLE
PS> New-UIPluginVersion -Plugin 1234 -File ./myapp.zip -Version "2.0.0"

Create a new version for a plugin


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

        # The ZIP file to be uploaded
        [Parameter()]
        [string]
        $File,

        # Plugin version (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Version,

        # List of tags associated to the version (required)
        [Parameter(Mandatory = $true)]
        [string[]]
        $Tags
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins versions create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Plugin `
            | Group-ClientRequests `
            | c8y ui plugins versions create $c8yargs
        }
        
    }

    End {}
}
