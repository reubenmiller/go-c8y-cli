# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserGroupCollection {
<#
.SYNOPSIS
Get user group collection

.DESCRIPTION
Get collection of (user) groups

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/usergroups_list

.EXAMPLE
PS> Get-UserGroupCollection

Get a list of user groups for the current tenant


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "usergroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.groupCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.group+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y usergroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y usergroups list $c8yargs
        }
    }

    End {}
}
