[cmdletbinding()]
Param()

. $PSScriptRoot/New-C8yPowershellApi.ps1
. $PSScriptRoot/New-C8yPowershellArguments.ps1
. $PSScriptRoot/New-C8yApiPowershellCommand.ps1
. $PSScriptRoot/New-C8yApiPowershellTest.ps1

#
# Use specs to generate powershell code
#
$BaseDir = Resolve-Path "$PSScriptRoot/../../tools/PSc8y"
$SpecFiles = Get-ChildItem -Path "$PSScriptRoot/../../api/spec/json" -Filter "*.json"

foreach ($iFile in $SpecFiles) {
    Write-Host ("Generating go cli code [{0}]" -f $iFile.Name) -ForegroundColor Gray
    New-C8yPowershellApi $iFile.FullName -OutputDir "$BaseDir/Public"
}

#
# Build the c8y cli binaries for each environment
#
& "$PSScriptRoot/../build-cli/build-binary.ps1" -OutputDir "$BaseDir/Dependencies" -All

# Build PowerShell Module
Import-Module "$PSScriptRoot/../../tools/PSc8y/tools/build.psm1" -Force
Export-ProductionModule

Write-Host "Build successful! $BaseDir"
