Function Register-Alias {
<#
.SYNOPSIS
Register aliases for commonly used cmdlets within the PSc8y module

.DESCRIPTION
Registers the aliases for quicker access to the PSc8y cmdlets.

Additional aliases can be created by using the in-built Powershell `New-Alias` cmdlet.

.EXAMPLE
Register-Alias

.LINK
Unregister-Alias
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

    # Add alias to c8y binary (only if the user does not already have access to it)
    if (-Not (Get-Command "c8y" -CommandType Application -ErrorAction SilentlyContinue)) {
        Set-Alias -Name "c8y" -Value (Get-ClientBinary) -Scope "Global"
    }
}
