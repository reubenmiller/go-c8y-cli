$script:CompleteTenantOptionCategory = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-TenantOptionCollection -PageSize 100 -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_.category -like "$searchFor*" } `
    | Select-Object -Unique -ExpandProperty category `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$script:CompleteTenantOptionKey = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    $options = Get-TenantOptionForCategory -Category $fakeBoundParameters["category"] -PageSize 100 -WarningAction SilentlyContinue -AsPSObject
    
    if ($options -is [hashtable]) {
        $keys = $options.keys
    } else {
        $keys = $options.psobject.Properties.Name
    }
    
    $keys `
    | Where-Object { $_ -like "$searchFor*" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}