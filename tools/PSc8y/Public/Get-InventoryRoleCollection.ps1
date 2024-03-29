﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-InventoryRoleCollection {
<#
.SYNOPSIS
Get inventory role collection

.DESCRIPTION
Get a list of inventory roles

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/users_listInventoryRoles

.EXAMPLE
PS> Get-InventoryRoleCollection

Get list of inventory roles


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users listInventoryRoles"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventoryrolecollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.inventoryrole+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y users listInventoryRoles $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y users listInventoryRoles $c8yargs
        }
    }

    End {}
}
