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
        $ConvertToPS = ($true -ne $BoundParameters["AsJson"]) `
            -and ($true -ne $BoundParameters["AsCSV"]) `
            -and ($true -ne $BoundParameters["Progress"])
        [PSCustomObject]@{
            ConvertToPS = $ConvertToPS
        }
    }   
}
