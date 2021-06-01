Function New-TemporaryDirectory {
<# 
.SYNOPSIS
Create a new temporary directory

.DESCRIPTION
Create a temporary directory in the systems temp directory folder.

.EXAMPLE
New-TemporaryDirectory

Create a new temp directory
#>
    [cmdletbinding()]
    Param()
    $parent = [System.IO.Path]::GetTempPath()
    [string] $name = [System.Guid]::NewGuid()
    New-Item -ItemType Directory -Path (Join-Path $parent $name)
}