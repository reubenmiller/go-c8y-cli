Function Get-ClientBinaryVersion {
<# 
.SYNOPSIS
Get the c8y client binary version

.DESCRIPTION
The c8y client binary version is the only dependency of the PSc8y module, and hence
the version number is helpful to determine what functions are available

.EXAMPLE
Get-ClientBinaryVersion

Show the client binary version on the console
#>
    [cmdletbinding()]
    Param()
    c8y version
}
