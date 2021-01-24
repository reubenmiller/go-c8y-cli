Function Register-ClientArgumentCompleter {
    <# 
.SYNOPSIS
Register PSc8y argument completions for a specific cmdlet

.DESCRIPTION
The cmdlet enables support for argument completion which are used within PSc8y in other modules.

.NOTES
The following argument completions are supports

* `-Session` - Session selection completion
* `-Template` - Template file completion

.EXAMPLE
Register-ClientArgumentCompleter -Name "Get-MyCustomCommand"

Register PSc8y argument completion for supported parameters for a custom function called "Get-MyCustomCommand" 

.EXAMPLE
Register-ClientArgumentCompleter -Name "New-CustomManagedObject" -Force

Force the registration of argument completers on a function which uses dynamic parameters
#>
    [cmdletbinding()]
    Param(
        # Command Name
        [Parameter(
            Mandatory = $true,
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true
        )]
        [string[]]
        $Name,

        # Force the registration of all parameter (required when a cmdlet has Dynamic Parameters)
        [switch]
        $Force
    )

    Process {
        foreach ($iCommand in (Get-Command $Name)) {

            if ($Force -or ($null -ne $iCommand.Parameters -and $iCommand.Parameters.ContainsKey("Session"))) {
                # Session
                Register-ArgumentCompleter -CommandName $iCommand -ParameterName Session -ScriptBlock {
                    param ($commandName, $parameterName, $wordToComplete)
                    $C8ySessionHome = Get-SessionHomePath
                    Get-ChildItem -Path $C8ySessionHome -Filter "$wordToComplete*.json" -ErrorAction SilentlyContinue -WarningAction SilentlyContinue | ForEach-Object {
                        [System.Management.Automation.CompletionResult]::new($_.BaseName, $_.BaseName, 'ParameterValue', $_.BaseName)
                    }
                }
            }

            # Template
            if ($Force -or ($null -ne $iCommand.Parameters -and $iCommand.Parameters.ContainsKey("Template"))) {
                Register-ArgumentCompleter -CommandName $iCommand -ParameterName Template -ScriptBlock {
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
            }
        }
    }
}
