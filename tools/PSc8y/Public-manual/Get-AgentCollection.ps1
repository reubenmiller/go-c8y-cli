Function Get-AgentCollection {
<#
.SYNOPSIS
Get a collection of agents

.DESCRIPTION
Get a collection of agent in the current tenant

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
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
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

        # Query.
        [Parameter(Mandatory = $false)]
        [string]
        $Query,

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
        Get-ClientCommonParameters -Type "Collection" -BoundParameters $PSBoundParameters
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            ,(c8y devices list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions)
        }
        else {
            c8y devices list $c8yargs
        }
    }
}
