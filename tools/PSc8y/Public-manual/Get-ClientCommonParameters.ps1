Function Get-ClientCommonParameters {
<#
.SYNOPSIS
Get the common parameters which can be added to a function which extends PSc8y functionality

.DESCRIPTION
* PageSize

.EXAMPLE
Function Get-MyObject {
    [cmdletbinding(
        SupportsShouldProcess = $True,
        ConfirmImpact = "None"
    )]
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
        [ValidateSet("Collection", "Get", "Create", "Update", "Delete", "Template", "")]
        [string[]]
        $Type
    )

    Process {
        $ParentCommand = @(Get-PSCallStack)[1].InvocationInfo.MyCommand

        $Dictionary = New-Object System.Management.Automation.RuntimeDefinedParameterDictionary
        foreach ($iType in $Type) {
            switch ($iType) {
                "Collection" {
                    New-DynamicParam -Name "PageSize" -Type "int" -DPDictionary $Dictionary
                    New-DynamicParam -Name "WithTotalPages" -Type "switch" -DPDictionary $Dictionary
                    New-DynamicParam -Name "CurrentPage" -Type "int" -DPDictionary $Dictionary
                    New-DynamicParam -Name "TotalPages" -Type "int" -DPDictionary $Dictionary
                    New-DynamicParam -Name "IncludeAll" -Type "switch" -DPDictionary $Dictionary
                }

                {$_ -match "Create|Update|Delete" } {
                    if ($_ -notmatch "Delete") {
                        New-DynamicParam -Name "Data" -Type "object" -DPDictionary $Dictionary
                    }
                    New-DynamicParam -Name "NoAccept" -Type "switch" -DPDictionary $Dictionary
                    New-DynamicParam -Name "ProcessingMode" -Type "string" -ValidateSet @("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "") -DPDictionary $Dictionary
                    New-DynamicParam -Name "Force" -Type "switch" -DPDictionary $Dictionary
                }

                "Template" {
                    New-DynamicParam -Name "Template" -Type "string" -DPDictionary $Dictionary
                    New-DynamicParam -Name "TemplateVars" -Type "string" -DPDictionary $Dictionary

                    # Completions
                    if ($ParentCommand) {
                        Register-ArgumentCompleter -CommandName $ParentCommand -ParameterName Template -ScriptBlock $script:CompletionTemplate
                    }
                }
            }
        }

        # Common parameters
        New-DynamicParam -Name Raw -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name OutputFile -Type "string" -DPDictionary $Dictionary
        New-DynamicParam -Name Proxy -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name NoProxy -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Session -Type "string" -DPDictionary $Dictionary
        New-DynamicParam -Name Timeout -Type "double" -DPDictionary $Dictionary

        # JSON parsing options
        New-DynamicParam -Name Output -Type "string" -ValidateSet @("json", "csv", "csvheader", "table") -DPDictionary $Dictionary
        New-DynamicParam -Name AsHashTable -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name AsPSObject -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Flatten -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Compress -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Pretty -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name NoColor -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Color -Type "switch" -DPDictionary $Dictionary

        # Confirmation
        New-DynamicParam -Name Prompt -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name ConfirmText -Type "string" -DPDictionary $Dictionary

        # Error options
        New-DynamicParam -Name WithError -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name SilentStatusCodes -Type "string" -DPDictionary $Dictionary

        # WhatIf options
        New-DynamicParam -Name Dry -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name DryFormat -Type "string" -ValidateSet @("markdown", "json", "dump", "curl") -DPDictionary $Dictionary
        # New-DynamicParam -Name WhatIfFormat -Type "string" -ValidateSet @("markdown", "json", "dump", "curl") -DPDictionary $Dictionary

        # Workers
        New-DynamicParam -Name Workers -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name Delay -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name MaxJobs -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name Progress -Type "switch" -DPDictionary $Dictionary

        # Activity logger
        New-DynamicParam -Name NoLog -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name LogMessage -Type "string" -DPDictionary $Dictionary

        # Select
        New-DynamicParam -Name Select -Type "string[]" -DPDictionary $Dictionary
        # New-DynamicParam -Name Filter -Type "string" -DPDictionary $Dictionary

        $Dictionary
    }
}
