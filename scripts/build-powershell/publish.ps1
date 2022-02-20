[cmdletbinding()]
Param()
$ErrorActionPreference = 'Stop'

Import-Module "$PSScriptRoot/../../tools/PSc8y/tools/build.psm1" -Force

# PowerShellGet 2.2.3 required to run correctly on MacOS
try {
	$PowerShellGetVersion = Get-Module -Name PowerShellGet -ListAvailable | ForEach-Object { [version] $_.Version } | Sort-Object -Descending | Select-Object -First 1

	if ($PowerShellGetVersion -lt ([version] "2.2.3")) {
		Install-Module PowerShellGet -MinimumVersion "2.2.3" -Force
		Remove-Module PowerShellGet -Force
		Start-Sleep -Seconds 2
		Import-Module PowerShellGet -MinimumVersion "2.2.3"
	}
} catch {
	Write-Host "PowerShellGet modules"
	Get-Module -Name PowerShellGet -ListAvailable

	$Versions = Get-Module -Name PowerShellGet | Select-Object -ExpandProperty Version
	Write-Host ("Current loaded version: {0}" -f ($Versions -join ","))
}


if ($env:APPVEYOR) {
	& $PSScriptRoot/wait-for-jobs.ps1
}

try {
	#
	# Build binaries
	#
	$ArtifactFolder = Export-ProductionModule
	$DependenciesDir = "$ArtifactFolder/Dependencies/"
	& $PSScriptRoot/../build-cli/build-binary.ps1 -OutputDir $DependenciesDir -All

	[array] $c8ybinaries = Get-ChildItem -Path $DependenciesDir -Filter "*c8y*"

	if ($c8ybinaries.Count -lt 4) {
		Write-Error "Failed to find all 4 c8y binaries"
		Exit 1
	}

	Write-Host "Publishing module from folder [$ArtifactFolder]"
	## Publish module to PowerShell Gallery
	$publishParams = @{
		Path        = $ArtifactFolder
		NuGetApiKey = $env:NUGET_API_KEY
		Verbose = $true
	}
	try {
		Publish-Module @publishParams
	} catch {
		# Ignore already published versions incase if there is a problem when publishing to PSGallery (it happens from time to time)
		if ($_.Exception.Message -like "current version '*' is already available") {
			Write-Warning ("Version has already been published. {0}" -f $_.Exception)
		} else {
			throw
		}
	}

} catch {
	Write-Error -Message $_.Exception.Message
	$host.SetShouldExit($LastExitCode)
}
