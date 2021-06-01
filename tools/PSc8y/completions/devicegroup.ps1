$script:CompleteDeviceGroup = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-DeviceGroupCollection -Name "$searchFor*" -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.name -like "$searchFor*" } `
    | Sort-Object { $_.name } `
    | ForEach-Object {
        $details = ("{0} ({1})" -f $_.name, $_.id)
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $details,
            'ParameterValue',
            $details
        )
    }
}