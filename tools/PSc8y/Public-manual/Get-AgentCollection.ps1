Function Get-AgentCollection {
<#
.SYNOPSIS
Get a collection of agents

.DESCRIPTION
Get a collection of agent in the current tenant

.LINK
c8y agents list

.EXAMPLE
Get-AgentCollection -Name *sensor*

Get all agents with "sensor" in their name

.EXAMPLE
Get-AgentCollection -Name *sensor* -Type *c8y_* -PageSize 100

Get the first 100 agents with "sensor" in their name and has a type matching "c8y_"

.EXAMPLE
Get-AgentCollection -Query "lastUpdated.date gt '2020-01-01T00:00:00Z'"

Get a list of agents which have been updated more recently than 2020-01-01

#>
    [cmdletbinding(PositionalBinding=$true, HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Agent name. Wildcards accepted
        [Parameter(Mandatory = $false)]
        [string]
        $Name,

        # Agent type.
        [Parameter(Mandatory = $false)]
        [string]
        $Type,

        # Agent fragment type.
        [Parameter(Mandatory = $false)]
        [string]
        $FragmentType,

        # Agent owner.
        [Parameter(Mandatory = $false)]
        [string]
        $Owner,

        # Availability.
        [Parameter(Mandatory = $false)]
        [ValidateSet("AVAILABLE", "UNAVAILABLE", "MAINTENANCE")]
        [string]
        $Availability,

        # LastMessageDateFrom - c8y_Availability.lastMessage filter
        [Parameter(Mandatory = $false)]
        [string]
        $LastMessageDateFrom,

        # LastMessageDateTo - c8y_Availability.lastMessage filter
        [Parameter(Mandatory = $false)]
        [string]
        $LastMessageDateTo,

        # Group.
        [Parameter(Mandatory = $false)]
        [string]
        $Group,

        # Query.
        [Parameter(Mandatory = $false)]
        [string]
        $Query,

        # QueryTemplate.
        [Parameter(Mandatory = $false)]
        [string]
        $QueryTemplate,

        # Order results by a specific field. i.e. "name", "_id desc" or "creationTime.date asc".
        [Parameter(Mandatory = $false)]
        [string]
        $OrderBy,

        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customAgentCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customAgent+json"
            BoundParameters = $PSBoundParameters
        }
    }
    DynamicParam {
        Get-ClientCommonParameters -Type "Collection"
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            ,(c8y agents list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions)
        }
        else {
            c8y agents list $c8yargs
        }
    }
}
