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
        $UseNativeOutput = $BoundParameters["AsJSON"] `
            -or $BoundParameters["IncludeAll"] `
            -or $BoundParameters["AsCSV"] `
            -or $BoundParameters["Progress"]
        [PSCustomObject]@{
            ConvertToPS = !$UseNativeOutput
        }
    }   
}
