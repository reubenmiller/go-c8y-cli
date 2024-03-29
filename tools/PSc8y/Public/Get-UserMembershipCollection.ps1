﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserMembershipCollection {
<#
.SYNOPSIS
get user membership collection

.DESCRIPTION
Get information about all groups that a user is a member of

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/users_listUserMembership

.EXAMPLE
PS> Get-UserMembershipCollection -Id $User.id

Get a list of groups that a user belongs to


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Tenant
        [Parameter()]
        [object]
        $Tenant
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users listUserMembership"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.groupReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.groupReference+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y users listUserMembership $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y users listUserMembership $c8yargs
        }
        
    }

    End {}
}
