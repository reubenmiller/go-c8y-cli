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
        [ValidateSet("Collection", "Create", "Template")]
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

                "Create" {
                    New-DynamicParam -Name "Data" -Type "object" -DPDictionary $Dictionary
                    New-DynamicParam -Name "ProcessingMode" -Type "string" -ValidateSet @("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "") -DPDictionary $Dictionary
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
        New-DynamicParam -Name TimeoutSec -Type "double" -DPDictionary $Dictionary

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
