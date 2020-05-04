[cmdletbinding()]
Param()

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
$OutputDir = "$PSScriptRoot/../../output"
& "$PSScriptRoot/build-binary.ps1" -OutputDir $OutputDir
$OutputDir = Resolve-Path $OutputDir


# Generate code completions
if ($IsMacOS) {
    $BinaryName = "c8y.macos"
} elseif ($IsLinux) {
    $BinaryName = "c8y.linux"
} else {
    $BinaryName = "c8y.windows.exe"
}

& "$OutputDir/$BinaryName" completion powershell > "$OutputDir/c8y.completion.ps1"
& "$OutputDir/$BinaryName" completion bash > "$OutputDir/c8y.completion.sh"

Write-Host "Build successful! $OutputDir"
