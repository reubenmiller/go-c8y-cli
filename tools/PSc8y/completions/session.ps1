$script:CompletionSession = {
    param ($commandName, $parameterName, $wordToComplete)
    Get-ChildItem -Path (Get-SessionHomePath) -Filter "$wordToComplete*.*" -ErrorAction SilentlyContinue -WarningAction SilentlyContinue `
    | Where-Object { $_ -match ".(json|yaml|yml|toml|env|properties)$" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_.BaseName, $_.BaseName, 'ParameterValue', $_.BaseName)
    }
}