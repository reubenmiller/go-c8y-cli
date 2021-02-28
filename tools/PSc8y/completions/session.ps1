$script:CompletionSession = {
    param ($commandName, $parameterName, $wordToComplete)
    Get-ChildItem -Path (Get-SessionHomePath) -Filter "$wordToComplete*.json" -ErrorAction SilentlyContinue -WarningAction SilentlyContinue | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_.BaseName, $_.BaseName, 'ParameterValue', $_.BaseName)
    }
}