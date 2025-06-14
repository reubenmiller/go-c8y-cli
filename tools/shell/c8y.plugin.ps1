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
