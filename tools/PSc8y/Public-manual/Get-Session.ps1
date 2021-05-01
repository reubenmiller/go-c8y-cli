Function Get-Session {
<#
.SYNOPSIS
Get the active Cumulocity Session

.DESCRIPTION
Get the details about the active Cumulocity session which is used by all cmdlets

.LINK
c8y sessions get

.EXAMPLE
Get-Session

Get the current Cumulocity session

.EXAMPLE
Get-Session -Show

Print the current session information (if set)

.OUTPUTS
None
#>
    [CmdletBinding()]
    Param(
        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # Only print the session information
        [switch]
        $Show
    )

    $c8yArgs = New-Object System.Collections.ArrayList

    if ($Session) {
        $null = $c8yArgs.AddRange(@("--session", $Session))
    }

    if ($Show) {
        c8y sessions get $c8yArgs
        return
    }

    # Convert session to powershell psobject
    $null = $c8yArgs.Add("--output=json")
    $sessionResponse = c8y sessions get $c8yArgs
    $data = $sessionResponse | ConvertFrom-Json

    if ($env:C8Y_SETTINGS_LOGGER_HIDESENSITIVE -eq "true") {
        $data | Add-PowershellType "cumulocity/session-hide-sensitive"
    } else {
        $data | Add-PowershellType "cumulocity/session"
    }
}
