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
        
        $AsJSON = if ($BoundParameters.ContainsKey("AsJSON")) { $BoundParameters["AsJSON"] } else { $false }

        $SelectedType = if ($ItemType) { $ItemType } else { $Type }
        if ($Raw) {
            $SelectedType = $Type
        }
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

                    ConvertFrom-Json -InputObject $item -Depth:$Depth -AsHashtable:$AsHashTable `
                    | Add-PowershellType -Type $SelectedType
                } else {
                    $item
                }
            }
        }
    }
}