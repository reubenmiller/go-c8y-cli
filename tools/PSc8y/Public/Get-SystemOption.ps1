﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-SystemOption {
<#
.SYNOPSIS
Get system option

.DESCRIPTION
Get a system option by category and key

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/systemoptions_get

.EXAMPLE
PS> Get-SystemOption -Category "system" -Key "version"

Get system option value


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # System Option category (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Category,

        # System Option key (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Key
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "systemoptions get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | Group-ClientRequests `
            | c8y systemoptions get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y systemoptions get $c8yargs
        }
        
    }

    End {}
}
