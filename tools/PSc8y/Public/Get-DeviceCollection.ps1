# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceCollection {
<#
.SYNOPSIS
Get device collection

.DESCRIPTION
Get a collection of devices based on filter parameters

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_list

.EXAMPLE
PS> Get-DeviceCollection -Name "sensor*" -Type myType

c8y devices list --name "sensor*" --type myType


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

        # Only include agents
        [Parameter()]
        [switch]
        $Agents,

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
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDevice+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Query `
            | Group-ClientRequests `
            | c8y devices list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y devices list $c8yargs
        }
        
    }

    End {}
}
