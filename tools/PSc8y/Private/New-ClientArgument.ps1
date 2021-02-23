Function New-ClientArgument {
    <# 
    .SYNOPSIS
    Run a Cumulocity client command using the c8y binary. Only intended for internal usage only
    
    .DESCRIPTION
    The command is a wrapper around the c8y binary which is used to send the rest request to Cumulocity.
    
    The result will also be parsed, and Powershell type information will be added to the result set, so
    only relevant information is shown.
    #>
    [cmdletbinding()]
    Param(
        # Parameters which should be passed to the c8y binary
        # The full parameter name should be used (i.e. --header, and not -H)
        [hashtable] $Parameters,

        # Command
        [string] $Command
    )

    Process {

        $c8yargs = New-Object System.Collections.ArrayList
        $BoundParameters = @{} + $Parameters

        # strip automatic variables
        $BoundParameters.Keys -match "(Verbose|WhatIf|Variable|Action|Confirm|Buffer|Debug|AsJSON|AsHashtable|AsCSV|Force|Color|Pretty)$" | ForEach-Object {
            $BoundParameters.Remove($_)
        }
        
        foreach ($iKey in $BoundParameters.Keys) {
            $Value = $BoundParameters[$iKey]
        
            foreach ($iValue in $Value) {
                if ("$Value" -notmatch "^$") {
                    $key = $iKey[0].ToString().ToLowerInvariant() + $iKey.SubString(1)
                    if ($Value -is [bool] -and $Value) {
                        $null = $c8yargs.AddRange(@("--${key}"))
                    }
                    else {
                        if ($key -eq "data" -or $Value -is [hashtable] -or $Value -is [PSCustomObject]) {
                            $ArgValue = ConvertTo-JsonArgument $Value
                            # due to cli parsing, data needs to be sent using "="
                            $null = $c8yargs.AddRange(@("--${key}", $ArgValue))
                        }
                        else {
                            if ($Value -match " ") {
                                # $null = $c8yargs.AddRange(@("--${key}", "$Value"))
                                $null = $c8yargs.Add("--${key}=`"$Value`"")
                            }
                            else {
                                $null = $c8yargs.Add("--${key}=$Value")
                            }
                        }
                    }
                }
            }
        }
        
        if ($WhatIfPreference) {
            $null = $c8yargs.Add("--dry")
        }
        
        # Always use verbose as information is extracted from it
        if ($VerbosePreference) {
            $null = $c8yargs.Add("--verbose")
        }

        if ($Parameters["WithTotalPages"]) {
            $null = $c8yargs.Add("--raw")
        }
        
        if ($Parameters["Color"] -or $Parameters["Pretty"]) {
            $null = $c8yargs.Add("--noColor=false")
        } elseif ($Parameters["NoColor"]) {
            $null = $c8yargs.Add("--noColor")
        }

        if ($Parameters["Pretty"]) {
            $null = $c8yargs.Add("--compress=false")
        }

        if ($Parameters["AsCSV"]) {
            $null = $c8yargs.Add("--csv")
        }
        
        if ($null -ne $Parameters["CurrentPage"]) {
            $null = $c8yargs.AddRange(@("--currentPage", $CurrentPage))
        }
        
        if ($null -ne $Parameters["TotalPages"]) {
            $null = $c8yargs.AddRange(@("--totalPages", $TotalPages))
        }
        
        # Include all pagination results
        if ($Parameters["IncludeAll"]) {
            # Write-Warning "IncludeAll operation is currently not implemented"
            $null = $c8yargs.Add("--includeAll")
        }
        
        $c8ycli = Get-ClientBinary
        Write-Verbose "binary: $c8ycli"
        Write-Verbose ("command: c8y $Command {0}" -f $c8yargs -join " ")
        $c8yargs
    }
}
