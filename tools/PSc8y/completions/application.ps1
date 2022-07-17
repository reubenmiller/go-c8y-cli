$script:CompleteApplication = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-ApplicationCollection -PageSize 100 -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_.id -like "$searchFor*" -or $_.name -like "$searchFor*" } `
    | ForEach-Object {
        $value = $_.name
        $details = ("{0} ({1})" -f $_.id, $_.name)
        [System.Management.Automation.CompletionResult]::new(
            $value,
            $details,
            'ParameterValue',
            $value
        )
    }
}