﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UIExtensionVersionCollection {
<#
.SYNOPSIS
Get extension version collection

.DESCRIPTION
Get a collection of extension versions by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_extensions_versions_list

.EXAMPLE
PS> Get-UIExtensionVersionCollection -Extension 1234

Get extension versions


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Extension
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Extension
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui extensions versions list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationVersionCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.applicationVersion+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Extension `
            | Group-ClientRequests `
            | c8y ui extensions versions list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Extension `
            | Group-ClientRequests `
            | c8y ui extensions versions list $c8yargs
        }
        
    }

    End {}
}
