Function Get-ClientOutputOption {
    [cmdletbinding()]
    Param(
        
        [Parameter(
            Mandatory = $true,
            Position = 0)]
        [hashtable]
        $BoundParameters
    )

    Process {
        $UseNativeOutput = $BoundParameters["Output"] `
            -or $BoundParameters["IncludeAll"] `
            -or $BoundParameters["CsvFormat"] `
            -or $BoundParameters["ExcelFormat"] `
            -or $BoundParameters["Progress"]

        $UsePowershellTypes = $BoundParameters["Select"]

        [PSCustomObject]@{
            ConvertToPS = !$UseNativeOutput
            UsePowershellTypes = $UsePowershellTypes
        }
    }   
}
