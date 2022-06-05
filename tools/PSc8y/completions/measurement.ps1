$script:CompleteMeasurementFragmentType = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if (!$fakeBoundParameters.ContainsKey("Device")) {
        return
    }

    $device = $fakeBoundParameters["Device"]

    Get-SupportedMeasurements -Device:$device -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_ -like "$searchFor*" } `
    | Sort-Object { $_ } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$script:CompleteMeasurementSeries = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if (!$fakeBoundParameters.ContainsKey("Device")) {
        return
    }

    $device = $fakeBoundParameters["Device"]
    $valueFragmentType = $fakeBoundParameters["valueFragmentType"]
    
    Get-SupportedSeries -Device:$device -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_ -like "$valueFragmentType.*" } `
    | Where-Object { $_ -like "*.$searchFor*" } `
    | ForEach-Object { "$_".Split(".")[-1] } `
    | Sort-Object { $_ } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$script:CompleteMeasurementFullSeries = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if (!$fakeBoundParameters.ContainsKey("Device")) {
        return
    }

    $device = $fakeBoundParameters["Device"]
    
    Get-SupportedSeries -Device:$device -WarningAction SilentlyContinue -AsPSObject `
    | Where-Object { $_ -like "$searchFor*" } `
    | Sort-Object { $_ } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}