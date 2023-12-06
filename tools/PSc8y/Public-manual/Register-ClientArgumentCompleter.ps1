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
Register-ClientArgumentCompleter -Name "New-CustomManagedObject"

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
        [object]
        $Command,

        # Bound parameters were completion is to be activated for
        [hashtable]
        $BoundParameters
    )

    Begin {
        $script:CompletionMapping = @{
            "Device" = $script:CompleteDevice
        }
    }

    Process {
        foreach ($name in $BoundParameters.Keys) {
            if ($script:CompletionMapping.ContainsKey($name)) {
                Register-ArgumentCompleter `
                    -CommandName $Command `
                    -ParameterName $name `
                    -Force
                    -ScriptBlock $script:CompletionMapping[$name]
            }
        }
    }
}
