# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ManagedObject {
<#
.SYNOPSIS
Get managed objects

.DESCRIPTION
Get an existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_get

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id

Get a managed object

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Get-ManagedObject

Get a managed object (using pipeline)

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id -WithParents

Get a managed object with parent references


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

        # include a flat list of all parents and grandparents of the given object
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory get"
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
            | c8y inventory get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y inventory get $c8yargs
        }
        
    }

    End {}
}
