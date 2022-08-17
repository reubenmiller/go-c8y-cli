[cmdletbinding()]
Param(
    # Filter (regex) to use when filtering the test file names
    [string] $TestFileFilter = ".+",

    [array] $Tag = $null,

    [array] $ExcludeTag = $null,

    # Filter out test file names to exclude from the test runner
    [string] $TestFileExclude = ""
)

# Change console encoding
$ConsoleEncodingBackup = $null
$CurrentEncodingName = [Console]::Out.Encoding.EncodingName
$RequiredEncodingName = [System.Text.Encoding]::UTF8.EncodingName

if ($CurrentEncodingName -ne $RequiredEncodingName) {
    Write-Host ("Current console encoding is not correct. Changing from [{0}] to [{1}]" -f @(
        $CurrentEncodingName,
        $RequiredEncodingName
    ))
    $ConsoleEncodingBackup = [Console]::Out
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
}

$OldConfirmPreference = $global:ConfirmPreference
$global:ConfirmPreference = "None"

if (!(Get-Module "Pester")) {
    Install-Module "Pester" -MinimumVersion "5.0.0" -MaximumVersion "5.100.0" -Repository PSGallery -Force
    Import-Module "Pester" -MinimumVersion "5.0.0" -MaximumVersion "5.100.0"
}

$originalLocaltion = Get-Location
Set-Location $PSScriptRoot

# Create the artifacts folder if not present
if (!(Test-Path -Path "./reports" )) {
    $null = New-Item -ItemType directory -Path "./reports"
}

$ReportOutput = Join-Path (Resolve-Path ".") -ChildPath "reports/Report_Pester.xml"

$PesterConfig = [PesterConfiguration]@{
    Run = @{
        Path = "./Tests"
        Exit = $false
        PassThru = $true
    }
    Output = @{
        Verbosity = "Diagnostic"
    }
    TestResult = @{
        Enabled = $true
        TestSuiteName = "powershell"
        OutputPath = $ReportOutput
        OutputFormat = "NUnitXml"
    }
}

if ($null -ne $Tag -or $null -ne $ExcludeTag) {
    $PesterConfig.Filter = @{}
    if ($Tag) {
        $PesterConfig.Filter.Tag = $Tag
    }

    if ($ExcludeTag) {
        $PesterConfig.Filter.ExcludeTag = $ExcludeTag
    }
}

. ./Tests/imports.ps1

$ModulePath = (Get-Module PSc8y).Path
Write-Host "Module path: $ModulePath"

# Disable activity logging by default
$Env:C8Y_SETTINGS_ACTIVITYLOG_ENABLED = "false"
$Env:C8Y_SETTINGS_DEFAULTS_DRYFORMAT = "json"

$env:SKIP_IMPORT = "true"
$result = Invoke-Pester -Configuration:$PesterConfig

if ($result.FailedCount -gt 0) {
    $null = Rename-item $ReportOutput -NewName "Failed_$($TestFile.BaseName)_Pester.xml" -Force -ErrorAction SilentlyContinue
    Write-Host ("Failed test: " -f $TestFile.Name) -ForegroundColor Red
}

# Delete any microservices still running
Get-MicroserviceCollection -PageSize 100 | Where-Object { $_.name -like "*testms*" } | Remove-Microservice -Force

$global:ConfirmPreference = $OldConfirmPreference

Set-Location $originalLocaltion.Path

if ($null -ne $ConsoleEncodingBackup) {
    Write-Verbose "Restoring original console encoding"
    [Console]::OutputEncoding = $ConsoleEncodingBackup.Encoding
}

exit $result.FailedCount
