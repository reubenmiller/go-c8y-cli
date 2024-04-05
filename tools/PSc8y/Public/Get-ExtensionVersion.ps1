# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ExtensionVersion {
<#
.SYNOPSIS
Get a specific version of an extension

.DESCRIPTION
Retrieve the selected version of an extension in your tenant. To select the version, use only the version or only the tag query parameter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_extensions_versions_get

.EXAMPLE
PS> Get-ExtensionVersion -Extension 1234 -Tag tag1

Get extension version by tag

.EXAMPLE
PS> Get-ExtensionVersion -Extension 1234 -Version 1.0

Get extension version by version name


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

        # The version field of the extension version
        [Parameter()]
        [string]
        $Version,

        # The tag of the extension version
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui extensions versions get"
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
            | c8y ui extensions versions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Extension `
            | Group-ClientRequests `
            | c8y ui extensions versions get $c8yargs
        }
        
    }

    End {}
}
