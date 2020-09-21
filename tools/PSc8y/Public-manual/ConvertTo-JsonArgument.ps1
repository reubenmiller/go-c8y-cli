Function ConvertTo-JsonArgument {
<# 
.SYNOPSIS
Convert a powershell hashtable/object to a json escaped string

.DESCRIPTION
Helper function is used when passing Powershell hashtable or PSCustomObjects to
the c8y binary. Before the c8y cli binary can accept it, it must be converted to json.

The necessary character escaping of literal backslashed `\` will be done automatically.

.EXAMPLE
ConvertTo-JsonArgument @{ myValue = "1" }

Converts the hashtable to an escaped json string

```json
{\"myValue\":\"1\"}
```
#>
    [cmdletbinding()]
    Param(
        # Input object to be converted
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
