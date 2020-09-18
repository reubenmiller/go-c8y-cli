Function Clear-Session {
<#
.SYNOPSIS
Clear the active Cumulocity Session

.DESCRIPTION
Clear the active Cumulocity Session

.EXAMPLE
Clear-Session

Clears the current Cumulocity session

.OUTPUTS
None
#>
    [CmdletBinding()]
    Param()
    Write-Verbose "Clearing cumulocity session"
    $env:C8Y_SESSION = ""
}
