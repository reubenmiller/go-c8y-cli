Function ConvertTo-JsonArgument {
<# 
.SYNOPSIS
Convert a powershell hashtable/object to a json escaped string

.DESCRIPTION
Helper function is used when passing Powershell hashtable or PSCustomObjects to
the c8y binary. Before the c8y cli binary can accept it, it must be converted to json.

The necessary character escaping of literal backslashed `\` will be done automatically.

If Data parameter is a file path then it is returned as is.

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

    # A change in powershell handling of quoting was introduced in Powershell >= 7.3
    # This complicates a few things...
    # See issue for more details: https://github.com/PowerShell/PowerShell/issues/18554
    $NeedsQuotes = ($null -eq $PSNativeCommandArgumentPassing) -or ($PSNativeCommandArgumentPassing -ne 'Standard')

    if ($Data -is [string] -or $data -is [System.IO.FileSystemInfo]) {
        if ($Data -and (Test-Path $Data)) {
            # Return path as is (and let c8y binary handle it)
            return $Data
        }
        # If string, then validate if json was provided
        
        try {
            $JSONArgs = @{
                InputObject = $Data
                ErrorAction = "Stop"
            }
            $DataObj = (ConvertFrom-Json @JSONArgs)
        } catch {
            # Return as is (and let c8y binary handle it)
            return $Data
        }
    } else {
        $DataObj = $Data
    }

    if ($NeedsQuotes) {
        $jsonRaw = (ConvertTo-Json $DataObj -Compress -Depth 100)
        $strArg = "{0}" -f ($jsonRaw -replace '(?<!\\)"', '\"')

        # Replace space with unicode char, as space can have console parsing problems
        $strArg = $strArg -replace " ", "\u0020"
    } else {
        # Note: replace \" with the unicode character to prevent intepretation errors on the command line
        $jsonRaw = (ConvertTo-Json $DataObj -Compress -Depth 100) -replace '\\"', '\u0022'
        $strArg = $jsonRaw
    }

    $strArg
}
