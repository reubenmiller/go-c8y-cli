# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SmartGroupCollection {
<#
.SYNOPSIS
List smart group collection

.DESCRIPTION
Get a collection of smart groups based on filter parameters

.LINK
c8y smartgroups list

.EXAMPLE
PS> Get-SmartGroupCollection

Get a list of smart groups

.EXAMPLE
PS> Get-SmartGroupCollection -Name "$Name*"

Get a list of smart groups with the names starting with 'myText'


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Filter by name
        [Parameter()]
        [string]
        $Name,

        # Filter by fragment type
        [Parameter()]
        [string]
        $FragmentType,

        # Filter by owner
        [Parameter()]
        [string]
        $Owner,

        # Filter by device query
        [Parameter()]
        [string]
        $DeviceQuery,

        # Filter by owner
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

        # Only include invisible smart groups
        [Parameter()]
        [switch]
        $OnlyInvisible,

        # Only include visible smart groups
        [Parameter()]
        [switch]
        $OnlyVisible,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "smartgroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Query `
            | Group-ClientRequests `
            | c8y smartgroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y smartgroups list $c8yargs
        }
        
    }

    End {}
}
