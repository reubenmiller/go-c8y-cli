Function Set-EnvironmentVariablesFromSession {
<#
.SYNOPSIS
Set environment variables based on the current Cumulocity session

.NOTES
If no session is active, then it will clear the environment variables

.EXAMPLE
Set-EnvironmentVariablesFromSession

.OUTPUTS
None
#>
    [cmdletbinding()]
    Param()

    $Session = Get-Session -ErrorAction SilentlyContinue 2>$null

    # reset any enabled side-effect commands
    $env:C8Y_SETTINGS_MODE_ENABLECREATE = ""
    $env:C8Y_SETTINGS_MODE_ENABLEUPDATE = ""
    $env:C8Y_SETTINGS_MODE_ENABLEDELETE = ""

    if ($null -eq $Session)
    {
        Write-Verbose "Clearing the Cumulocity environment variables"
        $Env:C8Y_URL = ""
        $Env:C8Y_BASEURL = ""
        $Env:C8Y_HOST = ""
        $Env:C8Y_TENANT = ""
        $Env:C8Y_USER = ""
        $Env:C8Y_USERNAME = ""
        $Env:C8Y_PASSWORD = ""
        $Env:C8Y_CREDENTIAL_COOKIES_0 = ""
        $Env:C8Y_CREDENTIAL_COOKIES_1 = ""
        $Env:C8Y_CREDENTIAL_COOKIES_2 = ""
        $Env:C8Y_CREDENTIAL_COOKIES_3 = ""
        $Env:C8Y_CREDENTIAL_COOKIES_4 = ""
        return
    }

    Write-Verbose "Setting Cumulocity environment variables"

    $Env:C8Y_URL = $Session.host;       # Used by @c8y/cli
    $Env:C8Y_BASEURL = $Session.host;
    $Env:C8Y_HOST = $Session.host;

    if ($Session.tenant) {
        $Env:C8Y_TENANT = $Session.tenant;
    } else {
        $CurrentTenant = Get-CurrentTenant
        $Env:C8Y_TENANT = $CurrentTenant.name
    }

    $Env:C8Y_USER = $Session.username;
    $Env:C8Y_USERNAME = $Session.username;

    # Don't set the password as it should have already been set in Test-ClientPassphrase
    if ([string]::IsNullOrEmpty($Env:C8Y_PASSWORD)) {
        $Env:C8Y_PASSWORD = $Session.password;
    }
}
