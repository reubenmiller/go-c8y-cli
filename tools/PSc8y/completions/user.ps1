$script:CompleteUser = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-UserCollection -PageSize 1000 -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_.id -like "$searchFor*" -or $_.email -like "$searchFor*" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $_.id,
            'ParameterValue',
            $_.id
        )
    }
}