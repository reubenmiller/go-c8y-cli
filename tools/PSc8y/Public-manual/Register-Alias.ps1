Function Register-Alias {
<#
.SYNOPSIS
Register aliases for commonly used cmdlets within the PSc8y module

.DESCRIPTION
Registers the aliases for quicker access to the PSc8y cmdlets.

Additional aliases can be created by using the in-built Powershell `New-Alias` cmdlet.

.EXAMPLE
Register-Alias

Register c8y aliases in PowerShell

.EXAMPLE
Unregister-Alias

Unregister c8y aliases in PowerShell
#>
    [cmdletbinding()]
    Param()

    $Aliases = $script:Aliases

    foreach ($Alias in $Aliases.Keys) {
        $Value = $Aliases[$Alias]

        if ($Value -is [string]) {
            Set-Alias -Name $Alias -Value $Aliases[$Alias] -Scope "Global"
        }
    }
}
