$script:CompletionTemplate = {
    param ($commandName, $parameterName, $wordToComplete)

    $settings = Get-ClientSetting
    $c8yTemplateHome = $settings."settings.template.path"
    if (!$c8yTemplateHome) {
        return
    }
    Get-ChildItem -Path $c8yTemplateHome -Recurse -Filter "$wordToComplete*" -ErrorAction SilentlyContinue -WarningAction SilentlyContinue |
    ForEach-Object {
        if ($_.Extension -match "(jsonnet)$") {
            [System.Management.Automation.CompletionResult]::new($_.Name, $_.Name, 'ParameterValue', $_.Name)
        }
    }
}
