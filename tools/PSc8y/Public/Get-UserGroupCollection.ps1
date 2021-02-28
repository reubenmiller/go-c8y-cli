# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserGroupCollection {
<#
.SYNOPSIS
Get collection of (user) groups

.DESCRIPTION
Get collection of (user) groups

.LINK
c8y userGroups list

.EXAMPLE
PS> Get-UserGroupCollection

Get a list of user groups for the current tenant


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "userGroups list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.groupCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.group+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y userGroups list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y userGroups list $c8yargs
        }
    }

    End {}
}
