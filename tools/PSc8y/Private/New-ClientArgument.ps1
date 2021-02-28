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
        [string] $Command,

        # List of names to exclude when parsing the parameters
        [string[]] $Exclude
    )

    Process {

        $c8yargs = New-Object System.Collections.ArrayList
        $BoundParameters = @{} + $Parameters

        # strip automatic variables
        $BoundParameters.Keys -match "(Verbose|WhatIf|Variable|Action|Confirm|Buffer|Debug|AsJSON|AsHashtable|AsCSV|AsCSVWithHeader|Force|Color|Pretty)$" | ForEach-Object {
            $BoundParameters.Remove($_)
        }

        # Exclude select keys
        if ($Exclude -and $Exclude.Count -gt 0) {
            foreach ($key in (@() + $BoundParameters.Keys)) {
                if ($Exclude -contains $key) {
                    $BoundParameters.Remove($key)
                }
            }
        }

        foreach ($iKey in $BoundParameters.Keys) {
            $Value = $BoundParameters[$iKey]

            if ($null -ne $Value) {
                $key = $iKey[0].ToString().ToLowerInvariant() + $iKey.SubString(1)

                switch ($Value) {
                    # boolean
                    { $Value -is [bool] -and $Value } {
                        $null = $c8yargs.AddRange(@("--${key}"))
                        break
                    }

                    { $Value -is [switch] } {
                        if ($Value) {
                            $null = $c8yargs.AddRange(@("--${key}"))
                        } else {
                            $null = $c8yargs.AddRange(@("--${key}=false"))
                        }
                        break
                    }

                    # json like values
                    { $key -eq "data" -or $Value -is [hashtable] -or $Value -is [PSCustomObject] } {
                        $ArgValue = ConvertTo-JsonArgument $Value
                        # due to cli parsing, data needs to be sent using "="
                        $null = $c8yargs.AddRange(@("--${key}", $ArgValue))
                        break
                    }

                    { $Value -is [array] } {
                        $items = Expand-Id $Value
                        if ($items.Count -eq 1) {
                            $null = $c8yargs.Add("--${key}=$($items -join ',')")
                            
                        } elseif ($items.Count -gt 1) {
                            $null = $c8yargs.Add("--${key}=`"$($items -join ',')`"")
                        }
                        break
                    }

                    { $Value -match " " -and ![string]::IsNullOrWhiteSpace($Value) } {
                        $null = $c8yargs.Add("--${key}=`"$Value`"")
                        break
                    }

                    default {
                        if (![string]::IsNullOrWhiteSpace($Value)) {
                            $null = $c8yargs.Add("--${key}=$Value")
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
        
        if ($Parameters["Color"]) {
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

        if ($Parameters["AsCSVWithHeader"]) {
            $null = $c8yargs.Add("--csv")
            $null = $c8yargs.Add("--csvHeader")
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
        ,$c8yargs
    }
}
