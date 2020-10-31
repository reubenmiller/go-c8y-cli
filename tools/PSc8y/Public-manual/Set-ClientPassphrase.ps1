Function Set-ClientPassphrase {
<#
.SYNOPSIS
Set the client passphrase which is used to encrypt sensitive Cumulocity session information.

.DESCRIPTION
The passphrase is used to encrypte sensitive information such as passwords and authorization cookies.

The passphrase is saved in an environment variable where it used for all c8y commands to decrypt
the sensitive information.

.EXAMPLE
Set-ClientPassphrase

Set the passphrase if it is not already set
#>
    [cmdletbinding()]
    Param()

    Process {
        # Check encryption
        $Binary = Get-ClientBinary
        $passphraseCheck = & $Binary sessions checkPassphrase --json

        if ($LASTEXITCODE -ne 0) {
            Write-Error "$encryptionInfo"
            return
        }

        $encryptionInfo = $passphraseCheck | ConvertFrom-Json

        # Save passphrase to env variable
        $env:C8Y_PASSPHRASE = $encryptionInfo.passphrase
        $env:C8Y_PASSPHRASE_TEXT = $encryptionInfo.secretText
    }
}