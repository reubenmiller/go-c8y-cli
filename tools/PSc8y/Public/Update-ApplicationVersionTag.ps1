# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-ApplicationVersionTag {
<#
.SYNOPSIS
Replace an application version's tags

.DESCRIPTION
Replaces the tags of a given application version in your tenant

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_versions_update

.EXAMPLE
PS> Update-ApplicationVersionTag -Id 1234 -Tag tag1

Get application version by tag

.EXAMPLE
PS> Update-ApplicationVersionTag -Id 1234 -Version 1.0

Get application version by version name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Application version
        [Parameter()]
        [string]
        $Version,

        # Tag assigned to the version. Version tags must be unique across all versions and version fields of application versions
        [Parameter()]
        [string[]]
        $Tag
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications versions update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationVersion+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y applications versions update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y applications versions update $c8yargs
        }
        
    }

    End {}
}
