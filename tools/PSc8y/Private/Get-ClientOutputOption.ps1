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
        $ConvertToPS = $BoundParameters["AsHashTable"] `
            -or $BoundParameters["AsPSObject"]
        $UsePowershellTypes = $BoundParameters["Select"]

        [PSCustomObject]@{
            ConvertToPS = $ConvertToPS
            UsePowershellTypes = $UsePowershellTypes
        }
    }   
}
