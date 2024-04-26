# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ApplicationVersion {
<#
.SYNOPSIS
Get a specific version of an application

.DESCRIPTION
Retrieve the selected version of an application in your tenant. To select the version, use only the version or only the tag query parameter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_versions_get

.EXAMPLE
PS> Get-ApplicationVersion -Application 1234 -Tag tag1

Get application version by tag

.EXAMPLE
PS> Get-ApplicationVersion -Application 1234 -Version 1.0

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
        $Application,

        # The version field of the application version
        [Parameter()]
        [string]
        $Version,

        # The tag of the application version
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications versions get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationVersion+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Application `
            | Group-ClientRequests `
            | c8y applications versions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Application `
            | Group-ClientRequests `
            | c8y applications versions get $c8yargs
        }
        
    }

    End {}
}
