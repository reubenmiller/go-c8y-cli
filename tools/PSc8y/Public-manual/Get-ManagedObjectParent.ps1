# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ManagedObjectParent {
<#
.SYNOPSIS
Get managed object parent

.DESCRIPTION
Get managed object parent

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_parents_get

.EXAMPLE
PS> Get-ManagedObjectParent -Id 1234 -Type addition

Get addition parent


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Type of relationship, e.g. addition, asset, device
        [Parameter()]
        [ValidateSet('addition','asset','device')]
        [string]
        $Type,

        # Number of parent jumps to do. 0 = current item, 1 = parent, 2 = grandparent etc. Defaults to 1. Use -1 for last parent
        [Parameter()]
        [int]
        $Level,

        # Return all parents in the chain
        [Parameter()]
        [switch]
        $All,

        # Return all parents in order from root to parent
        [Parameter()]
        [switch]
        $Reverse
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory parents get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y inventory parents get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y inventory parents get $c8yargs
        }
        
    }

    End {}
}
