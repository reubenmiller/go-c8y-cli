Function Unregister-Alias {
<#
.SYNOPSIS
Unregister aliases

.DESCRIPTION
Unregister any aliases which were registered by the Register-Alias cmdlet

.EXAMPLE
Unregister-Alias

.LINK
Register-Alias
#>
    [cmdletbinding()]
    Param()

    $Aliases = $script:Aliases

    foreach ($Alias in $Aliases.Keys) {
        $Value = $Aliases[$Alias]

        if ($Value -is [string]) {
            if (!(Get-Alias -Name $Value -ErrorAction SilentlyContinue)) {
                Remove-Item -Path Alias:$Alias
            }
        }
    }
}
