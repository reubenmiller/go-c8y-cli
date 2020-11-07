Function Test-ClientPassphrase {
<#
.SYNOPSIS
Test the client passphrase which is used to encrypt sensitive Cumulocity session information.

.DESCRIPTION
The passphrase is used to encrypt sensitive information such as passwords and authorization cookies.

The passphrase is saved in an environment variable where it used for all c8y commands to decrypt
the sensitive information.

.EXAMPLE
Test-ClientPassphrase

Set the passphrase if it is not already set
#>
    [cmdletbinding()]
    Param()

    Process {
        # Check encryption
        $Binary = Get-ClientBinary

        $c8yargs = New-object System.Collections.ArrayList
        $null = $c8yargs.AddRange(@(
            "sessions",
            "checkPassphrase",
            "--json"
        ))
        if ($VerbosePreference) {
            $null = $c8yargs.Add("--verbose")
        }
        $passphraseCheck = & $Binary $c8yargs

        if ($LASTEXITCODE -ne 0) {
            Write-Error "$encryptionInfo"
            return
        }

        $encryptionInfo = $passphraseCheck | ConvertFrom-Json

        # Save passphrase to env variable
        $env:C8Y_PASSWORD = $encryptionInfo.C8Y_PASSWORD
        $env:C8Y_PASSPHRASE = $encryptionInfo.C8Y_PASSPHRASE
        $env:C8Y_PASSPHRASE_TEXT = $encryptionInfo.C8Y_PASSPHRASE_TEXT
        $env:C8Y_CREDENTIAL_COOKIES_0 = $encryptionInfo.C8Y_CREDENTIAL_COOKIES_0
        $env:C8Y_CREDENTIAL_COOKIES_1 = $encryptionInfo.C8Y_CREDENTIAL_COOKIES_1
    }
}