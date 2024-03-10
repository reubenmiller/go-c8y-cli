# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ApplicationVersionCollection {
<#
.SYNOPSIS
Get application version collection

.DESCRIPTION
Get a collection of application versions by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_versions_list

.EXAMPLE
PS> Get-ApplicationVersionCollection -Id 1234

Get application versions


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
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications versions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationVersionCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.applicationVersion+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y applications versions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y applications versions list $c8yargs
        }
        
    }

    End {}
}
