# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-CurrentUser {
<#
.SYNOPSIS
Update the current user

.DESCRIPTION
Update properties or settings of your user such as first/last name, email or password


.EXAMPLE
PS> Update-CurrentUser -LastName "Smith"

Update the current user's lastname


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
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
        $Password,

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
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("FirstName")) {
            $Parameters["firstName"] = $FirstName
        }
        if ($PSBoundParameters.ContainsKey("LastName")) {
            $Parameters["lastName"] = $LastName
        }
        if ($PSBoundParameters.ContainsKey("Phone")) {
            $Parameters["phone"] = $Phone
        }
        if ($PSBoundParameters.ContainsKey("Email")) {
            $Parameters["email"] = $Email
        }
        if ($PSBoundParameters.ContainsKey("Enabled")) {
            $Parameters["enabled"] = $Enabled
        }
        if ($PSBoundParameters.ContainsKey("Password")) {
            $Parameters["password"] = $Password
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

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "users" `
                -Verb "updateCurrentUser" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.currentUser+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
