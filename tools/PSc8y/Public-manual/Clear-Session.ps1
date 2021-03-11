Function Clear-Session {
<#
.SYNOPSIS
Clear the active Cumulocity Session

.DESCRIPTION
Clear the active Cumulocity Session

.LINK
c8y sessions clear

.EXAMPLE
Clear-Session

Clears the current Cumulocity session

.OUTPUTS
None
#>
    [CmdletBinding()]
    Param()
    c8y sessions clear --shell powershell | Out-String | Invoke-Expression
}
