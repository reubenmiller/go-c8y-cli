# Code generated from specification version 1.0.0: DO NOT EDIT
Function Find-ManagedObjectCollection {
<#
.SYNOPSIS
Get a collection of managedObjects based on Cumulocity query language

.DESCRIPTION
Get a collection of managedObjects based on Cumulocity query language

.LINK
c8y inventory find

.EXAMPLE
PS> Find-ManagedObjectCollection -Query "name eq 'roomUpperFloor_*'"
Find all devices with their names starting with 'roomUpperFloor_'


#>
    [cmdletbinding(PositionalBinding=$true, HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject query. (required)
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Query,

        # String template to be used when applying the given query. Use %s to reference the query/pipeline input
        [Parameter(Mandatory = $false)]
        [string]
        $QueryTemplate,

        # ManagedObject sort results by.
        [Parameter(Mandatory = $false)]
        [string]
        $OrderBy,

        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents,

        # only include devices (i.e. add has(c8y_IsDevice) to the query)
        [Parameter()]
        [switch]        
        $OnlyDevices
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory find"
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
            | c8y inventory find $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y inventory find $c8yargs
        }
    }

    End {}
}
