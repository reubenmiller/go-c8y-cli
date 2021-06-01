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
        $BoundParameters.Keys -match "(Verbose|WhatIf|WhatIfFormat|Variable|Action|Buffer|Debug|AsHashtable|AsPSObject)$" | ForEach-Object {
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

                    { $Value -is [int] } {
                        $null = $c8yargs.AddRange(@("--${key}=$Value"))
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

        if ($DebugPreference -ne "SilentlyContinue") {
            $null = $c8yargs.Add("--debug")
        }

        if (-Not $Parameters.ContainsKey("Dry") -and $WhatIfPreference) {
            $null = $c8yargs.Add("--dry")
        }

        if (-Not $Parameters.ContainsKey("DryFormat") -and $Parameters["WhatIfFormat"]) {
            $null = $c8yargs.Add(("--dryFormat={0}" -f $Parameters["WhatIfFormat"]))
        }

        # Always use verbose as information is extracted from it
        if ($VerbosePreference) {
            $null = $c8yargs.Add("--verbose")
        }

        if ($Parameters["WithTotalPages"]) {
            $null = $c8yargs.Add("--raw")
        }

        # Allow empty pipes (only in powershell as it handles empty pipes in the cmdlets)
        # This simplifies logic on the powershell side, rather than dynamically checking if there is really piped
        # input or not
        $null = $c8yargs.Add("--allowEmptyPipe")

        ,$c8yargs
    }
}
