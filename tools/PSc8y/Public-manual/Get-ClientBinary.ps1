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

    if ($env:C8Y_BINARY)
    {
        Resolve-Path $env:C8Y_BINARY
        return
    }
}
