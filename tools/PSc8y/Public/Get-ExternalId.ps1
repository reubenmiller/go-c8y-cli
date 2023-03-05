﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ExternalId {
<#
.SYNOPSIS
Get external identity

.DESCRIPTION
Get an external identity object. An external identify will include the reference to a single device managed object


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/identity_get

.EXAMPLE
PS> Get-ExternalId -Type "my_SerialNumber" -Name "myserialnumber"

Get external identity


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # External identity type
        [Parameter()]
        [string]
        $Type,

        # External identity id/name (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "identity get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.externalId+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y identity get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y identity get $c8yargs
        }
        
    }

    End {}
}
