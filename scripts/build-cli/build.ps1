[cmdletbinding()]
Param(
    # Skip building the go binary. Only generate the go code
    [switch] $SkipBuildBinary
)

# Import functions
. $PSScriptRoot/New-C8yApi.ps1
. $PSScriptRoot/New-C8yApiGoCommand.ps1
. $PSScriptRoot/New-C8yApiGoRootCommand.ps1
. $PSScriptRoot/New-C8yApiGoGetValueFromFlag.ps1

#
# Generate go code from the specs
#
$OutputDir = Resolve-path (Join-Path $PSScriptRoot -ChildPath "../../pkg/cmd")

$SpecFiles = Get-ChildItem -Path "$PSScriptRoot/../../api/spec/json" -Filter "*.json"

$ImportStatements = foreach ($iFile in $SpecFiles) {
    Write-Host ("Generating go cli code [{0}]" -f $iFile.Name) -ForegroundColor Gray
    New-C8yApi $iFile.FullName -OutputDir $OutputDir
}
Write-Host "`nUse the following import statements in the root cmd`n"
$ImportStatements


#
# Build binary
#
if (-not $SkipBuildBinary) {
    $OutputDir = "$PSScriptRoot/../../output"
    & "$PSScriptRoot/build-binary.ps1" -OutputDir $OutputDir
    $OutputDir = Resolve-Path $OutputDir   
    Write-Host "Build successful! $OutputDir"
}
