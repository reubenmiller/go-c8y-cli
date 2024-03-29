﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserGroupMembershipCollection {
<#
.SYNOPSIS
Get users in group

.DESCRIPTION
Get all users in a user group

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/userreferences_listGroupMembership

.EXAMPLE
PS> Get-UserGroupMembershipCollection -Id $Group.id

List the users within a user group

.EXAMPLE
PS> Get-UserGroupByName -Name "business" | Get-UserGroupMembershipCollection

List the users within a user group (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group ID (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userreferences listGroupMembership"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.userReferenceCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.user+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y userreferences listGroupMembership $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y userreferences listGroupMembership $c8yargs
        }
        
    }

    End {}
}
