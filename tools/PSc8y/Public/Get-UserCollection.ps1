# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-UserCollection {
<#
.SYNOPSIS
Get a collection of users based on filter parameters

.DESCRIPTION
Get a collection of users based on filter parameters

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
        $Tenant,

        # Maximum number of results
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateRange(1,2000)]
        [int]
        $PageSize,

        # Include total pages statistic
        [Parameter()]
        [switch]
        $WithTotalPages,

        # Show the full (raw) response from Cumulocity including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Write the response to file
        [Parameter()]
        [string]
        $OutputFile,

        # Ignore any proxy settings when running the cmdlet
        [Parameter()]
        [switch]
        $NoProxy,

        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Username")) {
            $Parameters["username"] = $Username
        }
        if ($PSBoundParameters.ContainsKey("Groups")) {
            $Parameters["groups"] = $Groups
        }
        if ($PSBoundParameters.ContainsKey("Owner")) {
            $Parameters["owner"] = $Owner
        }
        if ($PSBoundParameters.ContainsKey("OnlyDevices")) {
            $Parameters["onlyDevices"] = $OnlyDevices
        }
        if ($PSBoundParameters.ContainsKey("WithSubusersCount")) {
            $Parameters["withSubusersCount"] = $WithSubusersCount
        }
        if ($PSBoundParameters.ContainsKey("WithApps")) {
            $Parameters["withApps"] = $WithApps
        }
        if ($PSBoundParameters.ContainsKey("WithGroups")) {
            $Parameters["withGroups"] = $WithGroups
        }
        if ($PSBoundParameters.ContainsKey("WithRoles")) {
            $Parameters["withRoles"] = $WithRoles
        }
        if ($PSBoundParameters.ContainsKey("Tenant")) {
            $Parameters["tenant"] = $Tenant
        }
        if ($PSBoundParameters.ContainsKey("PageSize")) {
            $Parameters["pageSize"] = $PageSize
        }
        if ($PSBoundParameters.ContainsKey("WithTotalPages") -and $WithTotalPages) {
            $Parameters["withTotalPages"] = $WithTotalPages
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

    }

    Process {
        foreach ($item in @("")) {


            Invoke-ClientCommand `
                -Noun "users" `
                -Verb "list" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.userCollection+json" `
                -ItemType "application/vnd.com.nsn.cumulocity.user+json" `
                -ResultProperty "users" `
                -Raw:$Raw
        }
    }

    End {}
}
