Function Get-ClientCommonParameters {
<#
.SYNOPSIS
Get the common parameters which can be added to a function which extends PSc8y functionality

.DESCRIPTION
* PageSize

.EXAMPLE
Function Get-MyObject {
    [cmdletbinding()]
    Param()

    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Process {
        Find-ManagedObjects @PSBoundParameters
    }
}
Inherit common parameters to a custom function. This will add parameters such as "PageSize", "TotalPages", "Template" to your function
#>
    [cmdletbinding()]
    Param(
        # Parameter types to include
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [ValidateSet("Collection", "Get", "Create", "Update", "Delete", "Template", "TemplateVars", "")]
        [string[]]
        $Type,

        # Ignore confirm parameter (i.e. when using inbuilt powershell -Confirm)
        [switch] $SkipConfirm
    )

    Process {
        $ParentCommand = @(Get-PSCallStack)[1].InvocationInfo.MyCommand

        $Dictionary = New-Object System.Management.Automation.RuntimeDefinedParameterDictionary
        foreach ($iType in $Type) {
            switch ($iType) {
                "Collection" {
                    New-DynamicParam -Name "PageSize" -Type "int" -DPDictionary $Dictionary -HelpMessage "Maximum results per page"
                    New-DynamicParam -Name "WithTotalPages" -Type "switch" -DPDictionary $Dictionary -HelpMessage "Request Cumulocity to include the total pages in the response statitics under .statistics.totalPages"
                    New-DynamicParam -Name "CurrentPage" -Type "int" -DPDictionary $Dictionary -HelpMessage "Current page which should be returned"
                    New-DynamicParam -Name "TotalPages" -Type "int" -DPDictionary $Dictionary -HelpMessage "Total number of pages to get"
                    New-DynamicParam -Name "IncludeAll" -Type "switch" -DPDictionary $Dictionary -HelpMessage "Include all results by iterating through each page"
                }

                {$_ -match "Create|Update|Delete" } {
                    if ($_ -notmatch "Delete") {
                        New-DynamicParam -Name "Data" -Type "object" -DPDictionary $Dictionary -HelpMessage "static data to be applied to body. accepts json or shorthande json, i.e. --data 'value1=1,my.nested.value=100'"
                    }
                    New-DynamicParam -Name "NoAccept" -Type "switch" -DPDictionary $Dictionary -HelpMessage "Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect"
                    New-DynamicParam -Name "ProcessingMode" -Type "string" -ValidateSet @("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "") -DPDictionary $Dictionary -HelpMessage "Cumulocity processing mode"
                    New-DynamicParam -Name "Force" -Type "switch" -DPDictionary $Dictionary -HelpMessage "Do not prompt for confirmation. Ignored when using --confirm"
                }

                "Template" {
                    New-DynamicParam -Name "Template" -Type "string" -DPDictionary $Dictionary -HelpMessage "Body template"
                    New-DynamicParam -Name "TemplateVars" -Type "string" -DPDictionary $Dictionary -HelpMessage "Body template variables"

                    # Completions
                    if ($ParentCommand) {
                        Register-ArgumentCompleter -CommandName $ParentCommand -ParameterName Template -ScriptBlock $script:CompletionTemplate
                    }
                }

                # Only template variables
                "TemplateVars" {
                    New-DynamicParam -Name "TemplateVars" -Type "string" -DPDictionary $Dictionary -HelpMessage "Body template variables"
                }
            }
        }

        # Common parameters
        New-DynamicParam -Name Raw -Type "switch" -DPDictionary $Dictionary -HelpMessage "Show raw response. This mode will force output=json and view=off"
        New-DynamicParam -Name OutputFile -Type "string" -DPDictionary $Dictionary -HelpMessage "Save JSON output to file (after select/view)"
        New-DynamicParam -Name OutputFileRaw -Type "string" -DPDictionary $Dictionary -HelpMessage "Save raw response to file (before select/view)"
        New-DynamicParam -Name Proxy -Type "switch" -DPDictionary $Dictionary -HelpMessage "Proxy setting, i.e. http://10.0.0.1:8080"
        New-DynamicParam -Name NoProxy -Type "switch" -DPDictionary $Dictionary -HelpMessage "Ignore the proxy settings"
        New-DynamicParam -Name Timeout -Type "string" -DPDictionary $Dictionary -HelpMessage "Request timeout. It accepts a duration, i.e. 1ms, 0.5s, 1m etc."
        
        # Session
        New-DynamicParam -Name Session -Type "string" -DPDictionary $Dictionary -HelpMessage "Session configuration"
        New-DynamicParam -Name SessionUsername -Type "string" -DPDictionary $Dictionary -HelpMessage "Override session username. i.e. peter or t1234/peter (with tenant)"
        New-DynamicParam -Name SessionPassword -Type "string" -DPDictionary $Dictionary -HelpMessage "Override session password"
        if ($ParentCommand) {
            Register-ArgumentCompleter -CommandName $ParentCommand -ParameterName Session -ScriptBlock $script:CompletionSession
        }

        # JSON parsing options
        New-DynamicParam -Name Output -Type "string" -ValidateSet @("json", "csv", "csvheader", "table", "serverresponse") -DPDictionary $Dictionary -HelpMessage "Output format i.e. table, json, csv, csvheader"
        New-DynamicParam -Name View -Type "string" -ValidateSet @("off", "auto") -DPDictionary $Dictionary -HelpMessage "Use views when displaying data on the terminal. Disable using --view off"
        New-DynamicParam -Name AsHashTable -Type "switch" -DPDictionary $Dictionary -HelpMessage "Return output as PowerShell Hashtables"
        New-DynamicParam -Name AsPSObject -Type "switch" -DPDictionary $Dictionary -HelpMessage "Return output as PowerShell PSCustomObjects"
        New-DynamicParam -Name Flatten -Type "switch" -DPDictionary $Dictionary -HelpMessage "flatten json output by replacing nested json properties with properties where their names are represented by dot notation"
        New-DynamicParam -Name Compact -Alias "Compress" -Type "switch" -DPDictionary $Dictionary -HelpMessage "Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal"
        New-DynamicParam -Name NoColor -Type "switch" -DPDictionary $Dictionary -HelpMessage "Don't use colors when displaying log entries on the console"
        
        # Help
        New-DynamicParam -Name Help -Type "switch" -DPDictionary $Dictionary -HelpMessage "Show command help"
        New-DynamicParam -Name Examples -Type "switch" -DPDictionary $Dictionary -HelpMessage "Show examples for the current command"

        # Confirmation
        if (-Not $SkipConfirm) {
            New-DynamicParam -Name Confirm -Type "switch" -DPDictionary $Dictionary -HelpMessage "Prompt for confirmation"
        }
        New-DynamicParam -Name ConfirmText -Type "string" -DPDictionary $Dictionary -HelpMessage "Custom confirmation text"

        # Error options
        New-DynamicParam -Name WithError -Type "switch" -DPDictionary $Dictionary -HelpMessage "Errors will be printed on stdout instead of stderr"
        New-DynamicParam -Name SilentStatusCodes -Type "string" -DPDictionary $Dictionary -HelpMessage "Status codes which will not print out an error message"

        # Dry options
        New-DynamicParam -Name Dry -Type "switch" -DPDictionary $Dictionary -HelpMessage "Dry run. Don't send any data to the server"
        New-DynamicParam -Name DryFormat -Type "string" -ValidateSet @("markdown", "json", "dump", "curl") -DPDictionary $Dictionary -HelpMessage "Dry run output format. i.e. json, dump, markdown or curl"

        # Workers
        New-DynamicParam -Name Workers -Type "int" -DPDictionary $Dictionary -HelpMessage "Number of workers"
        New-DynamicParam -Name Delay -Type "string" -DPDictionary $Dictionary -HelpMessage "delay after each request. It accepts a duration, i.e. 1ms, 0.5s, 1m etc."
        New-DynamicParam -Name DelayBefore -Type "string" -DPDictionary $Dictionary -HelpMessage "delay before each request. It accepts a duration, i.e. 1ms, 0.5s, 1m etc."
        New-DynamicParam -Name MaxJobs -Type "int" -DPDictionary $Dictionary -HelpMessage "Maximum number of jobs. 0 = unlimited (use with caution!)"
        New-DynamicParam -Name Progress -Type "switch" -DPDictionary $Dictionary -HelpMessage "Show progress bar. This will also disable any other verbose output"
        New-DynamicParam -Name AbortOnErrors -Type "int" -DPDictionary $Dictionary -HelpMessage "Abort batch when reaching specified number of errors"

        # Activity logger
        New-DynamicParam -Name NoLog -Type "switch" -DPDictionary $Dictionary -HelpMessage "Disables the activity log for the current command"
        New-DynamicParam -Name LogMessage -Type "string" -DPDictionary $Dictionary -HelpMessage "Add custom message to the activity log"

        # Select
        New-DynamicParam -Name Select -Type "string[]" -DPDictionary $Dictionary -HelpMessage "Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'"
        New-DynamicParam -Name Filter -Type "string[]" -DPDictionary $Dictionary -HelpMessage "Apply a client side filter to response before returning it to the user"

        # General rest options
        New-DynamicParam -Name Header -Type "string[]" -DPDictionary $Dictionary -HelpMessage "custom headers. i.e. --header 'Accept: value, AnotherHeader: myvalue'"
        New-DynamicParam -Name CustomQueryParam -Type "string[]" -DPDictionary $Dictionary -HelpMessage "add custom URL query parameters. i.e. --customQueryParam 'withCustomOption=true,myOtherOption=myvalue'"

        $Dictionary
    }
}
