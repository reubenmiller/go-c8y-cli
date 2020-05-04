Function ConvertTo-JsonArgument {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [object] $Data
    )

    if ($Data -is [string]) {
        # If string, then validate if json was provided
        $DataObj = (ConvertFrom-Json $Data)
    } else {
        $DataObj = $Data
    }

    # Note: replace \" with the unicode character to prevent intepretation errors on the command line
    $jsonRaw = (ConvertTo-Json $DataObj -Compress) -replace '\\"', '\u0022'
    $strArg = "{0}" -f ($jsonRaw -replace '(?<!\\)"', '\"')

    # Replace space with unicode char, as space can have console parsing problems
    $strArg = $strArg -replace " ", "\u0020"
    $strArg
}
