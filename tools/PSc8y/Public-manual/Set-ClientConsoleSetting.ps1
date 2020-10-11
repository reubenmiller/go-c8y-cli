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
Set-ClientConsoleSetting -EnableCreateCommands -EnableUpdateCommands

Enable all create and update commands until the session is changed
#>
    [cmdletbinding()]
    Param(
        # Hide all sensitive session information (tenant, username, password, base64 encoded passwords etc.)
        [switch] $HideSensitive,

        # Show sensitive information (excepts clear-text passwords)
        [switch] $ShowSensitive,

        # Enable create commands
        [switch] $EnableCreateCommands,

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
        $env:C8Y_LOGGER_HIDE_SENSITIVE = $false
    }

    if ($HideSensitive) {
        Write-Host "Sensitive session information will be hidden" -ForegroundColor Gray
        $env:C8Y_LOGGER_HIDE_SENSITIVE = $true
    }

    if ($DisableCommands) {
        Write-Host "Disableing create/update/delete commands" -ForegroundColor Gray
        $env:C8Y_SETTINGS_MODE_ENABLECREATE = ""
        $env:C8Y_SETTINGS_MODE_ENABLEUPDATE = ""
        $env:C8Y_SETTINGS_MODE_ENABLEDELETE = ""
    }

    if ($EnableCreateCommands) {
        Write-Host "Enabling create commands" -ForegroundColor Gray
        $env:C8Y_SETTINGS_MODE_ENABLECREATE = $true
    }

    if ($EnableUpdateCommands) {
        Write-Host "Enabling update commands" -ForegroundColor Gray
        $env:C8Y_SETTINGS_MODE_ENABLEUPDATE = $true
    }
    
    if ($EnableDeleteCommands) {
        Write-Host "Enabling delete commands" -ForegroundColor Gray
        $env:C8Y_SETTINGS_MODE_ENABLEDELETE = $true
    }

    if ($PSBoundParameters.ContainsKey("DefaultPageSize")) {
        if ($DefaultPageSize -gt 0) {
            $env:C8Y_DEFAULT_PAGESIZE = "$DefaultPageSize"
        } else {
            $env:C8Y_DEFAULT_PAGESIZE = ""
        }
    }
}
