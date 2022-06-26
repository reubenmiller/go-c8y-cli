[cmdletbinding()]
Param(
	# Skip running logic to test if a session is loaded or not
	# Useful when doing tests which rely on not having a session
	[switch] $SkipSessionTest
)

# Enable quick exit when running multiple tests
if ($env:SKIP_IMPORT -eq "true") {
	return
}

if (Get-Module PSc8y) {
	Remove-Module PSc8y -Force
}

Write-Verbose "PSScriptRoot: $PSScriptRoot";
#Import-Module Pester -MinimumVersion "5.0.0" -MaximumVersion "5.100.0"
$modulepath = Resolve-Path "$PSScriptRoot/../dist/PSc8y/PSc8y.psd1"
Import-Module $modulepath -Prefix "" -Force

# Import local functions which are only used in tests
. "$PSScriptRoot/../Public-manual/New-TestHostedApplication.ps1"
. "$PSScriptRoot/../Public-manual/New-TestMicroservice.ps1"

# Import helper functions
Get-ChildItem -Path "$PSScriptRoot/Helpers" -Recurse -Filter "*.ps1" | ForEach-Object {
	Write-Verbose ("Importing {0}" -f $_.FullName)
	. $_.FullName
}


# Add custom assertions
. "$PSScriptRoot/Assertions/ContainRequest.ps1"
Add-AssertionOperator -Name "ContainRequest" -Test $Function:ContainRequest -SupportsArrayInput -ErrorAction SilentlyContinue
. "$PSScriptRoot/Assertions/ContainInCollection.ps1"
Add-AssertionOperator -Name "ContainInCollection" -Test $Function:ContainInCollection -SupportsArrayInput -ErrorAction SilentlyContinue
. "$PSScriptRoot/Assertions/MatchObject.ps1"
Add-AssertionOperator -Name "MatchObject" -Test $Function:MatchObject -ErrorAction SilentlyContinue


# Get credentials from the environment
$env:C8Y_SETTINGS_CI = "true"

# required in non-interactive mode, otherwise powershell throws errors (regardless of confirmation preference)
$global:PSDefaultParameterValues = @{
	"*:Confirm" = $false;
	"*:Force" = $true;

	# Use PSoutput by default
	"*:AsPSObject" = $true;

	# required when using PowershellCore on linux
	# otherwise it will generate errors "You do not have sufficient access rights to perform this operation or the item is hidden, system, or read only."
	"Remove-Item:Force" = $true;
}

if (!$SkipSessionTest) {
	$TenantInfo = Get-CurrentTenant

	$User = Get-CurrentUser

	if (!$User) {
		Write-Error 'No Cumulocity Session found. Please set $env:C8Y_SESSION or $env:C8Y_HOST, $env:C8Y_TENANT, $env:C8Y_USER, $env:C8Y_PASSWORD and try again'
	}
	Write-Host ("Session: {0}/{1} on {2}" -f $TenantInfo.name, $User.id, $TenantInfo.domainName)
}
