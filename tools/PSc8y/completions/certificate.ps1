$script:CompleteDeviceCertificate = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-DeviceCertificateCollection -PageSize 100 -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_.fingerprint -like "$searchFor*" -or $_.name -like "$searchFor*" } `
    | ForEach-Object {
        $value = $_.name
        $details = ("{0} ({1})" -f $_.fingerprint, $_.name)
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $details,
            'ParameterValue',
            $_.id
        )
    }
}