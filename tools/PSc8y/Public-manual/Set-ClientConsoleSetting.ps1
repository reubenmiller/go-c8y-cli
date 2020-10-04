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
#>
    [cmdletbinding()]
    Param(
        # Hide all sensitive session information (tenant, username, password, base64 encoded passwords etc.)
        [switch] $HideSensitive,

        # Show sensitive information (excepts clear-text passwords)
        [switch] $ShowSensitive,

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

    if ($PSBoundParameters.ContainsKey("DefaultPageSize")) {
        if ($DefaultPageSize -gt 0) {
            $env:C8Y_DEFAULT_PAGESIZE = "$DefaultPageSize"
        } else {
            $env:C8Y_DEFAULT_PAGESIZE = ""
        }
    }
}
