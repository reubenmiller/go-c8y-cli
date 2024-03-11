# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ApplicationVersion {
<#
.SYNOPSIS
Delete a specific version of an application

.DESCRIPTION
Delete a specific version of an application in your tenant, by a given tag or version

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_versions_delete

.EXAMPLE
PS> Remove-ApplicationVersion -Application 1234 -Tag tag1

Delete application version by tag

.EXAMPLE
PS> Remove-ApplicationVersion -Application 1234 -Version 1.0

Delete application version by version name


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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications versions delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Application `
            | Group-ClientRequests `
            | c8y applications versions delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Application `
            | Group-ClientRequests `
            | c8y applications versions delete $c8yargs
        }
        
    }

    End {}
}
