Function Import-ClientBinary {
    <#
.SYNOPSIS
Import/Download the Cumulocity Binary

.DESCRIPTION
Get the full path to the Cumulocity Binary which is compatible with the current Operating system

.EXAMPLE
Import-ClientBinary

Download the client binary corresponding to your current platform
#>
    [cmdletbinding()]
    [OutputType([String])]
    Param(
        # Version. Defaults to the module's version
        [string] $Version,

        [ValidateSet("macOS", "windows", "linux")]
        [string] $Platform,

        [ValidateSet("amd64", "386", "arm64", "armv5")]
        [string] $Arch
    )

    # Version
    if ([string]::IsNullOrWhiteSpace($Version)) {
        $Version = "" + $MyInvocation.MyCommand.ScriptBlock.Module.Version

        if ($Version -eq "0.0") {
            $ModuleDefinition = Join-Path $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase "PSc8y.psd1"
            $RawVersion = Get-Content $ModuleDefinition | Where-Object { $_ -like "ModuleVersion*=*" } | Select-Object -First 1
            if ($RawVersion -match "'(\S+)'") {
                $Version = $Matches[1]
            }
        }
    }

    # Platform
    if ([string]::IsNullOrWhiteSpace($Platform)) {
        if ($IsLinux) {
            $Platform = "linux"
        }
        elseif ($IsMacOS) {
            $Platform = "macOS"
        }
        else {
            $Platform = "windows"
        }
    }

    # Architecture
    if ([string]::IsNullOrWhiteSpace($Arch)) {
        switch ($Platform) {
            "windows" {
                $Arch = switch ($Env:PROCESSOR_ARCHITECTURE) {
                    "amd64" { "amd64" }
                    "x86" { "386" }
                    "x64" { "amd64" }
                }
            }

            "macOS" {
                $Arch = switch -regex (uname -m) {
                    "^armv8l|armv8b|arm64|aarch64$" { "arm64" }
                    "^x86_64$" { "amd64" }
                    "^i386$" { "386" }
                }
            }

            "linux" {
                $Arch = switch -regex (uname -m) {
                    "^armv8l|armv8b|arm64|aarch64$" { "arm64" }
                    "^arm$" { "armv5" }
                    "^x86_64$" { "amd64" }
                    "^i386$" { "386" }
                }
            }
        }
    }

    switch ($Platform) {
        "windows" {
            $FileExtension = "zip"
            $BinaryExtension = ".exe"
        }
        default {
            $FileExtension = "tar.gz"
            $BinaryExtension = ""
        }
    }

    $BinaryName = "c8y_${Version}_${Platform}_${Arch}$BinaryExtension"
    $TargetFile = "c8y_${Version}_${Platform}_${Arch}.${FileExtension}"
    $LocalBinary = Join-Path $script:Dependencies $BinaryName
    $DownloadUrl = "https://github.com/reubenmiller/go-c8y-cli/releases/download/v$Version/$TargetFile"

    if (-Not (Test-Path $script:Dependencies)) {
        $null = New-Item $script:Dependencies -ItemType Directory
    }

    try {
        if (-Not (Test-Path $LocalBinary)) {
            $TempFile = New-TemporaryFile

            Invoke-RestMethod -Uri $DownloadUrl -OutFile $TempFile

            Write-Verbose "Downloading binary from $DownloadUrl to $TempFile"

            switch ($FileExtension) {
                "tar.gz" {
                    $TempDir = New-TemporaryDirectory
                    $null = tar -C $TempDir --strip-components 1 -xzvf $TempFile "c8y_${Version}_${Platform}_${Arch}/bin/c8y" 2> $null

                    Move-Item "$TempDir/bin/c8y" $LocalBinary -Force
                    Remove-Item $TempFile
                    Remove-Item $TempDir -Recurse
                }
                "zip" {
                    $TempDir = New-TemporaryDirectory
                    Expand-Archive $TempFile -DestinationPath $TempDir
                    Copy-Item "$TempDir/bin/c8y.exe" $LocalBinary -Force
                    Remove-Item $TempDir -Recurse
                }
            }
        }

        if ($IsMacOS -or $IsLinux) {
            chmod +x $LocalBinary
        }

        $env:C8Y_BINARY = $LocalBinary
        Resolve-Path $LocalBinary
    }
    catch {
        Write-error "Failed to download c8y binary. url=$DownloadUrl, error=$_"
    }
}
