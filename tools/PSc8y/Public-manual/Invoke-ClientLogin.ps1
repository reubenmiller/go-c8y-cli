Function Invoke-ClientLogin {
    [cmdletbinding()]
    Param(
        # Two Factor Authentication code
        [string] $TFACode
    )
    Process {
        $c8yBinary = Get-ClientBinary

        $cliArgs = New-Object System.Collections.ArrayList

        $null = $cliArgs.AddRange(@("sessions", "login"))

        if ($VerbosePreference) {
            $cliArgs.Add("--verbose")
        }

        $result = & $c8yBinary $cliArgs

        if ($LASTEXITCODE -ne 0) {
            return
        }

        $result
    }
}