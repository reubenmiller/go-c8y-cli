﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-Configuration {
<#
.SYNOPSIS
Get Configuration

.DESCRIPTION
Get an existing configuration package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/configuration_get

.EXAMPLE
PS> Get-Configuration -Id $mo.id

Get a configuration package

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Get-Configuration

Get a configuration package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Configuration package (managedObject) id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "configuration get"
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
            | c8y configuration get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y configuration get $c8yargs
        }
        
    }

    End {}
}
