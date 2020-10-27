Function Invoke-ClientLogin {
    [cmdletbinding()]
    Param(
        # Two Factor Authentication code
        [string] $TFACode,

        # Clear existing cookies (if present)
        [switch] $Clear
    )
    Process {
        $c8yBinary = Get-ClientBinary

        $cliArgs = New-Object System.Collections.ArrayList

        $null = $cliArgs.AddRange(@("sessions", "login"))

        if ($TFACode) {
            $null = $cliArgs.AddRange(@("--tfaCode", $TFACode))
        }

        if ($Clear) {
            $null = $cliArgs.AddRange(@("--clear"))
        }

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