$script:CompleteTenant = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if ($env:C8Y_TENANT) {
        [System.Management.Automation.CompletionResult]::new(
            $env:C8Y_TENANT,
            "(Current tenant)",
            'ParameterValue',
            $env:C8Y_TENANT
        )
    }

    Get-TenantCollection -PageSize 100 -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_.id -like "$searchFor*" } `
    | ForEach-Object {
        $id = $_.id
        $details = ("{0} ({1})" -f $_.id, ($_.domain -split "\.")[0])
        [System.Management.Automation.CompletionResult]::new(
            $id,
            $details,
            'ParameterValue',
            $id
        )
    }
}