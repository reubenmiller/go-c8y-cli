Function Get-SessionHomePath {
<# 
.SYNOPSIS
Get session home path

.DESCRIPTION
Get the path where all sessions are stored

.Link
c8y settings list

.EXAMPLE
Get-SessionHomePath 
#>
    [cmdletbinding()]
    Param()
    c8y settings list --select "session.home" --csv
}
