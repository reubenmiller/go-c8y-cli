﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-TenantOptionCollection {
<#
.SYNOPSIS
Get tenant option collection

.DESCRIPTION
Get collection of tenant options

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenantoptions_list

.EXAMPLE
PS> Get-TenantOptionCollection

Get a list of tenant options


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantoptions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.optionCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.option+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y tenantoptions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y tenantoptions list $c8yargs
        }
    }

    End {}
}
