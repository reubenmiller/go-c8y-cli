Function Set-ClientConsoleSetting {
<#
.SYNOPSIS
Set console settings to be used by the cli tool

.DESCRIPTION
Sensitive information:
When using -HideSensitive, the following information will be obfuscated when shown on the console
(tenant, username, password, base64 credentials)

.EXAMPLE
Set-ClientConsoleSetting -HideSensitive

Hide any sensitive session information on the console. Settings like (tenant, username, password, base64 credentials)

.EXAMPLE
Set-ClientConsoleSetting -Mode qual

Enable all create and update commands until the session is changed
#>
    [cmdletbinding()]
    Param(
        # Hide all sensitive session information (tenant, username, password, base64 encoded passwords etc.)
        [switch] $HideSensitive,

        # Show sensitive information (excepts clear-text passwords)
        [switch] $ShowSensitive,

        # Session mode
        [string] $Mode,

        # Enable update commands
        [switch] $EnableUpdateCommands,

        # Enable delete commands
        [switch] $EnableDeleteCommands,

        # Disable all create/update/delete commands
        [switch] $DisableCommands,

        # Set the default paging size to use in collection queries
        [int] $DefaultPageSize
    )

    if ($ShowSensitive) {
        Write-Host "Sensitive session information will be visible (except clear-text passwords)" -ForegroundColor Gray
        $env:C8Y_SETTINGS_LOGGER_HIDESENSITIVE = $false
    }

    if ($HideSensitive) {
        Write-Host "Sensitive session information will be hidden" -ForegroundColor Gray
        $env:C8Y_SETTINGS_LOGGER_HIDESENSITIVE = $true
    }

    if ($DisableCommands) {
        Write-Host "Disabling create/update/delete commands" -ForegroundColor Gray
        $env:C8Y_MODE = "prod"
    }

    if ($Mode) {
        Write-Host "Setting session mode" -ForegroundColor Gray
        $env:C8Y_MODE = $Mode
    }

    if ($PSBoundParameters.ContainsKey("DefaultPageSize")) {
        if ($DefaultPageSize -gt 0) {
            $env:C8Y_SETTINGS_DEFAULTS_PAGESIZE = "$DefaultPageSize"
        } else {
            $env:C8Y_SETTINGS_DEFAULTS_PAGESIZE = ""
        }
    }
}
