[cmdletbinding()]
Param(
    # Filter (regex) to use when filtering the test file names
    [string] $TestFileFilter = ".+",

    # Filter out test file names to exclude from the test runner
    [string] $TestFileExclude = "",

    # Throttle number of concurrent tests (grouped by test file)
    [int] $ThrottleLimit = 10
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
} else {
    # Remove existing reports
    Remove-Item "./reports/*xml"
}

# Dot source the invoke-parallel script
# . "$PSScriptRoot/tools/Invoke-Parallel.ps1"

$Tests = Get-ChildItem "./Tests" -Filter "*.tests.ps*" -Recurse |
    Where-Object { $_.Name -match "$TestFileFilter" } | 
    Where-Object {
        if ($TestFileExclude) {
            $_.Name -notmatch "$TestFileExclude"
        } else {
            $true
        }
    } |
    Foreach-Object {
        @{
            File = $_
        }
    }

$TestStartTime = Get-Date

# $results = $Tests | Invoke-Parallel -Throttle 5 -ScriptBlock {
$results = $Tests | ForEach-Object -ThrottleLimit:$ThrottleLimit -Parallel {
    $TestItem = $_
    $TestFile = $TestItem.File

    $ConfirmPreference = "None"

    Write-Host ("Starting file: {0}" -f $TestFile.Name) -ForegroundColor Gray

    $ReportOutput = Join-Path (Resolve-Path ".") -ChildPath "reports/Report_$($TestFile.BaseName)_Pester.xml"

    $PesterConfig = [PesterConfiguration]@{
        Run = @{
            Path = $TestFile.FullName
            Exit = $false
            PassThru = $true
        }
        Output = @{
            Verbosity = "Diagnostic"
        }
        TestResult = @{
            Enabled = $true
            TestSuiteName = $TestFile.Name
            OutputPath = $ReportOutput
            OutputFormat = "NUnitXml"
        }
    }

    . ./Tests/imports.ps1

    # Disable activity logging by default
    $Env:C8Y_SETTINGS_ACTIVITYLOG_ENABLED = "false"

    $result = Invoke-Pester -Configuration:$PesterConfig
    
    if ($result.FailedCount -gt 0) {
        $null = Rename-item $ReportOutput -NewName "Failed_$($TestFile.BaseName)_Pester.xml" -Force -ErrorAction SilentlyContinue
        Write-Host ("Failed test: " -f $TestFile.Name) -ForegroundColor Red
    }

    $result

    Write-Host ("Finished file: {0} - Failed count {1}" -f $TestFile.Name, $result.FailedCount) -ForegroundColor Gray
}

$TotalDuration = (Get-Date) - $TestStartTime

# Delete any microservices still running
Get-MicroserviceCollection -PageSize 100 | Where-Object { $_.name -like "*testms*" } | Remove-Microservice -Force

$totalSeconds = 0
$totalCount = 0
$failedCount = 0
$skippedCount = 0

$results | ForEach-Object {
    $totalCount += $_.TotalCount
    $failedCount += $_.FailedCount
    $skippedCount += $_.SkippedCount
    $totalSeconds += $_.Duration.TotalSeconds
}

$global:ConfirmPreference = $OldConfirmPreference
    
$code = $failedCount

$colour = "Green"

if ($failedCount -gt 0) {
    $colour = "Red"
    Write-Host "`nSome tests failed: " -NoNewLine -ForegroundColor:$colour
} else {
    Write-Host "`nAll tests passed: " -NoNewLine -ForegroundColor:$colour
}

Write-Host ("failed={0}, skipped={1}, total={2}, duration={3} s" -f @(
    $failedCount,
    $skippedCount,
    $totalCount,
    $TotalDuration.TotalSeconds
)) -ForegroundColor:$colour

Set-Location $originalLocaltion.Path

if ($null -ne $ConsoleEncodingBackup) {
    Write-Verbose "Restoring original console encoding"
    [Console]::OutputEncoding = $ConsoleEncodingBackup.Encoding
}

exit $code
