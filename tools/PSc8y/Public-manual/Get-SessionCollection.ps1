Function Get-SessionCollection {
<#
.SYNOPSIS
Get a collection of Cumulocity Sessions

.DESCRIPTION
Get a collection of Cumulocity Sessions found in the home folder under .cumulocity

.EXAMPLE
Get-SessionCollection

List all of the Cumulocity sessions in the default home folder

.OUTPUTS
object[]
#>
    [CmdletBinding()]
    Param()
    $HomePath = Get-SessionHomePath

    $JSONArgs = @{}
    if ($PSVersionTable.PSVersion.Major -gt 5) {
        $JSONArgs.Depth = 100
    }

    $Sessions = Get-ChildItem -LiteralPath $HomePath -Filter "*.json" -Recurse | ForEach-Object {
        $Path = $PSItem.FullName
        $data = Get-Content -LiteralPath $Path | ConvertFrom-Json @JSONArgs
        $data | Add-Member -MemberType NoteProperty -Name "path" -Value $Path -ErrorAction SilentlyContinue
        $data.path = $Path
        $data
    }

    $Sessions `
        | Select-Object `
        | Add-PowershellType "cumulocity/sessionCollection"
}
