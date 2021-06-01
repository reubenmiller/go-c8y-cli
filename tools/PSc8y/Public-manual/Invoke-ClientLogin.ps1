Function Invoke-ClientLogin {
    [cmdletbinding()]
    Param(
        # Two Factor Authentication code
        [string] $TFACode,

        # Clear existing token (if present)
        [switch] $Clear
    )
    Process {
        $cliArgs = New-Object System.Collections.ArrayList

        $null = $cliArgs.AddRange(@("--shell", "powershell"))

        if ($TFACode) {
            $null = $cliArgs.AddRange(@("--tfaCode", $TFACode))
        }

        if ($Clear) {
            $null = $cliArgs.AddRange(@("--clear"))
        }

        if ($VerbosePreference) {
            $cliArgs.Add("--verbose")
        }

        $result = c8y sessions login $cliArgs

        if ($LASTEXITCODE -ne 0) {
            return
        }

        $result | Out-String | Invoke-Expression
    }
}