# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-InventoryRole {
<#
.SYNOPSIS
Get inventory role

.DESCRIPTION
Get a specific inventory role

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/users_getInventoryRole

.EXAMPLE
PS> Get-InventoryRoleCollection -PageSize 1 | Get-InventoryRole

Get an inventory role (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Role id. Note: lookup by name is not yet supported (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users getInventoryRole"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventoryrole+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y users getInventoryRole $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y users getInventoryRole $c8yargs
        }
        
    }

    End {}
}
