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
        Get-ClientCommonParameters -Type "Create", "Template" -BoundParameters $PSBoundParameters
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
        $Type,

        # Existing bound parameters from the cmdlet. Providing it will ensure that the dynamic parameters do not duplicate
        # existing parameters.
        [Parameter()]
        [AllowNull()]
        [hashtable]
        $BoundParameters
    )
    
    Process {
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
                    New-DynamicParam -Name NoAccept -Type "switch" -DPDictionary $Dictionary
                    New-DynamicParam -Name "ProcessingMode" -Type "string" -ValidateSet @("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "") -DPDictionary $Dictionary
                    New-DynamicParam -Name "Force" -Type "switch" -DPDictionary $Dictionary
                }

                "Template" {
                    New-DynamicParam -Name "Template" -Type "string" -DPDictionary $Dictionary
                    New-DynamicParam -Name "TemplateVars" -Type "string" -DPDictionary $Dictionary
                }
            }
        }

        # Common parameters
        New-DynamicParam -Name Raw -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name OutputFile -Type "string" -DPDictionary $Dictionary
        New-DynamicParam -Name NoProxy -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Session -Type "string" -DPDictionary $Dictionary
        New-DynamicParam -Name Timeout -Type "double" -DPDictionary $Dictionary
        
        # JSON parsing options
        New-DynamicParam -Name Depth -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name AsHashTable -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name AsJSON -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name AsCSV -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Compress -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Pretty -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name NoColor -Type "switch" -DPDictionary $Dictionary
        New-DynamicParam -Name Color -Type "switch" -DPDictionary $Dictionary

        # Workers
        New-DynamicParam -Name Workers -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name Delay -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name MaxJobs -Type "int" -DPDictionary $Dictionary
        New-DynamicParam -Name Progress -Type "switch" -DPDictionary $Dictionary
        
        # Select
        New-DynamicParam -Name Select -Type "string" -DPDictionary $Dictionary
        # New-DynamicParam -Name Filter -Type "string" -DPDictionary $Dictionary

        # Remove key that are already defined
        if ($BoundParameters) {
            $BoundParameters.Keys | Foreach-Object {
                if ($Dictionary.ContainsKey($_)) {
                    $null = $Dictionary.Remove($_)
                }
            }
        }

        $Dictionary
    }
}
