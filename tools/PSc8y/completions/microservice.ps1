$script:CompleteMicroservice = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-ApplicationCollection -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.type -eq "MICROSERVICE" } `
    | Where-Object { $_.id -like "$searchFor*" -or $_.name -like "$searchFor*" } `
    | ForEach-Object {
        $details = ("{0} ({1})" -f $_.id, $_.name)
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $details,
            'ParameterValue',
            $_.id
        )
    }
}