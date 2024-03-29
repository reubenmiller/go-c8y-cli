﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RoleCollection {
<#
.SYNOPSIS
Get role collection

.DESCRIPTION
Get collection of user roles

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/userroles_list

.EXAMPLE
PS> Get-RoleCollection -PageSize 100

Get a list of roles


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userroles list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.roleCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.role+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y userroles list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y userroles list $c8yargs
        }
    }

    End {}
}
