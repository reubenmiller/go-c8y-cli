Function ConvertFrom-ClientOutput {
    [CmdletBinding()]
    param (
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0,
            Mandatory = $true
        )]
        [object[]]
        $InputObject,

        [string]
        $Type = "application/json",

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
            -or $BoundParameters["Raw"]
        
        $AsJSON = $BoundParameters["AsJSON"] `
            -or $BoundParameters["Pretty"] `
            -or $BoundParameters["Compress"] `
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
            if ($AsJSON) {
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