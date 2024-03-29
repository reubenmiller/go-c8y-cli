﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-CurrentUser {
<#
.SYNOPSIS
Get current user

.DESCRIPTION
Get the user representation associated with the current credentials used by the request

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/currentuser_get

.EXAMPLE
PS> Get-CurrentUser

Get the current user


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(

    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "currentuser get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.currentUser+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y currentuser get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y currentuser get $c8yargs
        }
    }

    End {}
}
