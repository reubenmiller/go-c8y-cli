$ErrorActionPreference = 'Stop'

try {
	Write-Host ("Powershell Version: {0}" -f $PSVersionTable.PSVersion.ToString())

	Import-Module -Name Pester
	$ProjectRoot = "$ENV:APPVEYOR_BUILD_FOLDER/tools/PSc8y"

	$testResultsFilePath = "$ProjectRoot/TestResults.xml"

	$invPesterParams = @{
        Script = "$ProjectRoot/Tests"
		OutputFormat = 'NUnitXml'
		OutputFile = $testResultsFilePath
		EnableExit = $true
		PassThru = $true
	}
	$results = Invoke-Pester @invPesterParams

    if ($env:APPVEYOR) {
        $Address = "https://ci.appveyor.com/api/testresults/nunit/$($env:APPVEYOR_JOB_ID)"
        (New-Object 'System.Net.WebClient').UploadFile( $Address, $testResultsFilePath )
	}

	if ($results.FailedCount -gt 0) {
		exit $results.FailedCount
	}

} catch {
	Write-Error -Message $_.Exception.Message
	$host.SetShouldExit($LastExitCode)
}
