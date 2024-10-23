[cmdletbinding()]
Param()

if (Get-Command "c8y" -ErrorAction SilentlyContinue) {
    c8y completion powershell | Out-String | Invoke-Expression
}

# PSReadline settings
Set-PSReadLineOption -EditMode Windows
Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete
# Set-PSReadLineKeyHandler -Key Tab -Function Complete
# Autocompletion for arrow keys
Set-PSReadlineKeyHandler -Key UpArrow -Function HistorySearchBackward
Set-PSReadlineKeyHandler -Key DownArrow -Function HistorySearchForward

########################################################################
# c8y helpers
########################################################################

Function Set-Session {
<#
.SYNOPSIS
Switch Cumulocity session interactively

.EXAMPLE
Set-Session myhost

Set session and only show session matching "myhost"
#>
    [cmdletbinding()]
    Param()

    $c8yenv = c8y sessions set --noColor=false $args
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "Set session failed"
        return
    }
    $c8yenv | Out-String | Invoke-Expression
}

Function Clear-Session {
<#
.SYNOPSIS
Clear all cumulocity session variables

.EXAMPLE
Clear-Session

Clear session variables
#>
    [cmdletbinding()]
    Param()
    c8y sessions clear | Out-String | Invoke-Expression
}

Function clear-c8ypassphrase {
<#
.SYNOPSIS
Clear the encryption passphrase environment variables

.EXAMPLE
clear-c8ypassphrase

Clear encryption passphrase variables
#>
    [cmdletbinding()]
    Param()
    $env:C8Y_PASSPHRASE = $null
    $env:C8Y_PASSPHRASE_TEXT = $null
}

Function set-c8ymode {
<#
.SYNOPSIS
Enable a c8y temporary mode by setting the environment variables

.EXAMPLE
set-c8ymode dev

Enable dev mode (enables POST, PUT and DELETE commands)
#>
    [cmdletbinding()]
    Param(
        # Mode
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [ValidateSet("dev", "qual", "prod")]
        [string]
        $Mode,

        [parameter(ValueFromRemainingArguments=$true)]
        $Options
    )
    c8y settings update --shell auto mode $Mode $Options | Out-String | Invoke-Expression
    Write-Host "Enabled "$Mode" mode (temporarily)" -ForegroundColor Green
}
