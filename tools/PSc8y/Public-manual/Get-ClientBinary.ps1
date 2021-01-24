Function Get-ClientBinary {
<#
.SYNOPSIS
Get the path to the Cumulocity Binary

.DESCRIPTION
Get the full path to the Cumulocity Binary which is compatible with the current Operating system

.EXAMPLE
Get-ClientBinary

Returns the fullname of the path to the Cumulocity binary
#>
    [cmdletbinding()]
    [OutputType([String])]
    Param()

    if ($IsLinux) {
        Resolve-Path (Join-Path $script:Dependencies "c8y.linux")
    } elseif ($IsMacOS) {
        Resolve-Path (Join-Path $script:Dependencies "c8y.macos")
    } else {
        Resolve-Path (Join-Path $script:Dependencies "c8y.windows.exe")
    }
}
