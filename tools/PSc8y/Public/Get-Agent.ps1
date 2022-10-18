# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-Agent {
<#
.SYNOPSIS
Get agent

.DESCRIPTION
Get an agent's managed object representation

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/agents_get

.EXAMPLE
PS> Get-Agent -Id $agent.id

Get agent by id

.EXAMPLE
PS> Get-Agent -Id $agent.name

Get agent by name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Agent ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved
        [Parameter()]
        [switch]
        $SkipChildrenNames,

        # Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.
        [Parameter()]
        [switch]
        $WithChildren,

        # When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)
        [Parameter()]
        [switch]
        $WithChildrenCount,

        # When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.
        [Parameter()]
        [switch]
        $WithGroups,

        # Include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customAgent+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y agents get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y agents get $c8yargs
        }
        
    }

    End {}
}
