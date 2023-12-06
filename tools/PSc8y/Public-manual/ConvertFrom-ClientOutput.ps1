Function ConvertFrom-ClientOutput {
    <# 
.SYNOPSIS
Convert the text output to PowerShell objects.

.DESCRIPTION
The cmdlet is used internally to interface between the c8y binary and PowerShell.

.EXAMPLE
c8y devices list | ConvertFrom-ClientOutput -Type mycustomtype

Convert the json output from the c8y devices list command into powershell objects

#>
    [CmdletBinding()]
    param (
        # Input object
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0,
            Mandatory = $true
        )]
        [object[]]
        $InputObject,

        # Type
        [string]
        $Type = "application/json",

        # Item type
        [string]
        $ItemType = "application/json",

        # Existing bound parameters from the cmdlet. Common parameters will be retrieved from it
        [Parameter()]
        [AllowNull()]
        [hashtable]
        $BoundParameters
    )

    Begin {
        $Depth = if ($BoundParameters.ContainsKey("Depth")) { $BoundParameters["Depth"] } else { 100 }
        $AsHashTable = if ($BoundParameters.ContainsKey("AsHashTable")) { $BoundParameters["AsHashTable"] } else { $false }
        $Raw = $BoundParameters["WithTotalPages"] `
            -or $BoundParameters["WithTotalElements"] `
            -or $BoundParameters["Raw"]
        
        $UseNativeOutput = $BoundParameters["AsJSON"] `
            -or $BoundParameters["Output"] -eq "csv" `
            -or $BoundParameters["Output"] -eq "csvheader" `
            -or $BoundParameters["Output"] -eq "table" `
            -or $BoundParameters["CsvFormat"] `
            -or $BoundParameters["ExcelFormat"] `
            -or $BoundParameters["Pretty"] `
            -or $BoundParameters["Compress"] `
            -or $BoundParameters["Compact"] `
            -or $BoundParameters["Dry"] `
            -or $BoundParameters["Help"] `
            -or $BoundParameters["Examples"] `
            -or $WhatIfPreference

        $SelectedType = if ($ItemType) { $ItemType } else { $Type }
        if ($Raw) {
            $SelectedType = $Type
        }

        # Ignore powershell type when using custom select properties, otherwise the user might not see the properties they want on the console
        $IgnorePowershellType = $BoundParameters["Select"].Count -gt 0
    }

    Process {
        foreach ($item in $InputObject) {
            if ($UseNativeOutput) {
                $item
            }
            else {
                # Detect json responses automatically
                if ($SelectedType -match "json") {
                    # Strip color codes (if present)
                    $item = $item -replace '\x1b\[[0-9;]*m'

                    $output = ConvertFrom-Json -InputObject $item -Depth:$Depth -AsHashtable:$AsHashTable `

                    if (-Not $IgnorePowershellType) {
                        $output = $output | Add-PowershellType -Type $SelectedType
                    }                
                    $output
                } else {
                    $item
                }
            }
        }
    }
}