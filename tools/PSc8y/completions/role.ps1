$script:CompleteRole = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-RoleCollection -PageSize 100 `
    | Select-Object -ExpandProperty id `
    | Where-Object { $_ -like "$searchFor*" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            "$_"
        )
    }
}