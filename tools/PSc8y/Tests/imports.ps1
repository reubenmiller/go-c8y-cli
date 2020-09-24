﻿
Remove-Module PSc8y -ErrorAction SilentlyContinue

Write-Verbose "PSScriptRoot: $PSSScriptRoot";
Import-Module Pester -MinimumVersion "5.0.0" -MaximumVersion "5.100.0"
Import-Module "$PSScriptRoot/../dist/PSc8y/PSc8y.psd1" -Prefix "" -Force

# Import local functions which are only used in tests
. "$PSScriptRoot/../Public-manual/New-TestHostedApplication.ps1"
. "$PSScriptRoot/../Public-manual/New-TestMicroservice.ps1"

# Import helper functions
. "$PSScriptRoot/Get-JSONFromResponse.ps1"

# Get credentials from the environment
$env:C8Y_USE_ENVIRONMENT = "on"

# required in non-interactive mode, otherwise powershell throws errors (regardless of confirmation preference)
$PSDefaultParameterValues = @{
	"*:Confirm" = $false;
	"*:Force" = $true;
}

$TenantInfo = Get-CurrentTenant

$User = Get-CurrentUser

if (!$User) {
    Write-Error 'No Cumulocity Session found. Please set $env:C8Y_SESSION or $env:C8Y_HOST, $env:C8Y_TENANT, $env:C8Y_USER, $env:C8Y_PASSWORD and try again'
}
Write-Host ("Session: {0}/{1} on {2}" -f $TenantInfo.name, $User.id, $TenantInfo.domainName)
