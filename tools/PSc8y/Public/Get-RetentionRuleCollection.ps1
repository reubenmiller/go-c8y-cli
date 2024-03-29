﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RetentionRuleCollection {
<#
.SYNOPSIS
Get retention rule collection

.DESCRIPTION
Get a collection of retention rules configured in the current tenant


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/retentionrules_list

.EXAMPLE
PS> Get-RetentionRuleCollection

Get a list of retention rules


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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "retentionrules list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.retentionRuleCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.retentionRule+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y retentionrules list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y retentionrules list $c8yargs
        }
    }

    End {}
}
