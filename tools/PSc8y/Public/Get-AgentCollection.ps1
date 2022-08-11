# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-AgentCollection {
<#
.SYNOPSIS
Get agent collection

.DESCRIPTION
Get a collection of agents based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/agents_list

.EXAMPLE
PS> Get-AgentCollection -Name "sensor*" -Type myType

Get a collection of agents with type "myType", and their names start with "sensor"


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Additional query filter
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Query,

        # String template to be used when applying the given query. Use %s to reference the query/pipeline input
        [Parameter()]
        [string]
        $QueryTemplate,

        # Order by. e.g. _id asc or name asc or creationTime.date desc
        [Parameter()]
        [string]
        $OrderBy,

        # Filter by name
        [Parameter()]
        [string]
        $Name,

        # Filter by type
        [Parameter()]
        [string]
        $Type,

        # Filter by fragment type
        [Parameter()]
        [string]
        $FragmentType,

        # Filter by owner
        [Parameter()]
        [string]
        $Owner,

        # Filter by c8y_Availability.status
        [Parameter()]
        [ValidateSet('AVAILABLE','UNAVAILABLE','MAINTENANCE')]
        [string]
        $Availability,

        # Filter c8y_Availability.lastMessage to a specific date
        [Parameter()]
        [string]
        $LastMessageDateTo,

        # Filter c8y_Availability.lastMessage from a specific date
        [Parameter()]
        [string]
        $LastMessageDateFrom,

        # Filter creationTime.date to a specific date
        [Parameter()]
        [string]
        $CreationTimeDateTo,

        # Filter creationTime.date from a specific date
        [Parameter()]
        [string]
        $CreationTimeDateFrom,

        # Filter by group inclusion
        [Parameter()]
        [object[]]
        $Group,

        # Include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedobjectcollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.customAgent+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Query `
            | Group-ClientRequests `
            | c8y agents list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y agents list $c8yargs
        }
        
    }

    End {}
}
