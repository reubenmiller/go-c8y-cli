Function Get-Session {
<#
.SYNOPSIS
Get the active Cumulocity Session

.DESCRIPTION
Get the details about the active Cumulocity session which is used by all cmdlets

.EXAMPLE
Get-Session

Get the current Cumulocity session

.OUTPUTS
None
#>
    [CmdletBinding()]
    Param(
        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session
    )

    $c8yBinary = Get-ClientBinary
    $c8yArgs = New-Object System.Collections.ArrayList
    $null = $c8yArgs.AddRange(@("sessions", "get", "--pretty=false"))

    if ($Session) {
        $null = $c8yArgs.AddRange(@("--session", $Session))
    }
    
    $sessionResponse = & $c8yBinary $c8yArgs 2>$null

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to get session details. reason: $sessionResponse"
        return
    }

    $JSONArgs = @{}
    if ($PSVersionTable.PSVersion.Major -gt 5) {
        $JSONArgs.Depth = 100
    }

    $data = $sessionResponse | ConvertFrom-Json @JSONArgs

    if ($env:C8Y_LOGGER_HIDE_SENSITIVE -eq "true") {
        $data | Add-PowershellType "cumulocity/session-hide-sensitive"
    } else {
        $data | Add-PowershellType "cumulocity/session"
    }
}
