[cmdletbinding()]
Param()
$OldConfirmPreference = $global:ConfirmPreference
$global:ConfirmPreference = "None"

if (!(Get-Module "Pester")) {
    Install-Module "Pester" -MinimumVersion "5.0.0" -MaximumVersion "5.100.0" -Repository PSGallery -Force
    Import-Module "Pester" -MinimumVersion "5.0.0" -MaximumVersion "5.100.0"
}

Set-Location $PSScriptRoot

# Create the artifacts folder if not present
if (!(Test-Path -Path "./reports" )){ $null = New-Item -ItemType directory -Path "./reports"}

# Dot source the invoke-parallel script
# . "$PSScriptRoot/tools/Invoke-Parallel.ps1"

$Tests = Get-ChildItem "./Tests" -Filter "*.tests.ps*"

$ThrottleLimit = 5

$TestStartTime = Get-Date

# $results = $Tests | Invoke-Parallel -Throttle 5 -ScriptBlock {
$results = $Tests | ForEach-Object -ThrottleLimit:$ThrottleLimit -Parallel {
    Write-Host "Invoking Pester for: $_.Name"

    $ConfirmPreference = "None"

    Write-Host ("Starting file: {0}" -f $_.Name) -ForegroundColor Gray

    $result = Invoke-Pester `
        -Script $_.FullName `
        -EnableExit `
        -Quiet `
        -PassThru `
        -OutputFile "./reports/Test-$($_.Name)_Pester.xml" `
        -OutputFormat NUnitXml
    
    if ($result.FailedCount -gt 0) {
        Rename-item "./reports/Test-$($_.Name)_Pester.xml" -NewName "Test-$($_.Name)_Pester.Failed.xml"
    }
    
    $result

    Write-Host ("Finished file: {0}" -f $_.Name) -ForegroundColor Gray
}

$TotalDuration = (Get-Date) - $TestStartTime


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

exit $code
