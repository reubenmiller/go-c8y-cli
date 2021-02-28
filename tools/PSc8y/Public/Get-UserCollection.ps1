# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserCollection {
<#
.SYNOPSIS
Get a collection of users based on filter parameters

.DESCRIPTION
Get a collection of users based on filter parameters

.LINK
c8y users list

.EXAMPLE
PS> Get-UserCollection

Get a list of users


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # prefix or full username
        [Parameter()]
        [string]
        $Username,

        # numeric group identifiers separated by commas; result will contain only users which belong to at least one of specified groups
        [Parameter()]
        [string]
        $Groups,

        # exact username
        [Parameter()]
        [string]
        $Owner,

        # If set to 'true', result will contain only users created during bootstrap process (starting with 'device_'). If flag is absent (or false) the result will not contain 'device_' users.
        [Parameter()]
        [switch]
        $OnlyDevices,

        # if set to 'true', then each of returned users will contain additional field 'subusersCount' - number of direct subusers (users with corresponding 'owner').
        [Parameter()]
        [switch]
        $WithSubusersCount,

        # Include applications related to the user
        [Parameter()]
        [switch]
        $WithApps,

        # Include group information
        [Parameter()]
        [switch]
        $WithGroups,

        # Include role information
        [Parameter()]
        [switch]
        $WithRoles,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.userCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.user+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y users list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y users list $c8yargs
        }
    }

    End {}
}
