# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-CurrentUser {
<#
.SYNOPSIS
Update current user

.DESCRIPTION
Update properties or settings of your user such as first/last name, email or password


.LINK
c8y users updateCurrentUser

.EXAMPLE
PS> Update-CurrentUser -LastName "Smith"

Update the current user's last name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # User first name
        [Parameter()]
        [string]
        $FirstName,

        # User last name
        [Parameter()]
        [string]
        $LastName,

        # User phone number. Format: '+[country code][number]', has to be a valid MSISDN
        [Parameter()]
        [string]
        $Phone,

        # User email address
        [Parameter()]
        [string]
        $Email,

        # User activation status (true/false)
        [Parameter()]
        [ValidateSet('true','false')]
        [string]
        $Enabled,

        # User password. Min: 6, max: 32 characters. Only Latin1 chars allowed
        [Parameter()]
        [string]
        $Password
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "users updateCurrentUser"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.currentUser+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y users updateCurrentUser $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y users updateCurrentUser $c8yargs
        }
    }

    End {}
}
