Function Install-ClientBinary {
<# 
.SYNOPSIS
Install the Cumulocity cli binary (c8y)

.DESCRIPTION
Install the Cumulocity cli binary (c8y) so it is accessible from everywhere in consoles (assuming /usr/local/bin is in the $PATH variable)

.EXAMPLE
Install-ClientBinary

On Linux/MacOS, this installs the cumulocity binary to /usr/local/bin
On Windows this will throw a warning

.EXAMPLE
Install-ClientBinary -InstallPath /usr/bin

Install the Cumulocity binary to /usr/bin

#>
    [cmdletbinding()]
    Param(
        # Cumulocity installation path where the c8y binaries will be installed. Defaults to $env:C8Y_INSTALL_PATH
        [Parameter(
            Position = 0
        )]
        [string] $InstallPath = $env:C8Y_INSTALL_PATH
    )

    $binary = Get-ClientBinary

    if (!$binary -or !(Test-Path $binary)) {
        Write-Error "Could not find c8y binary"
        return
    }

    if ($IsMacOS -or $IsLinux) {
        if (!$InstallPath) {
            $InstallPath = "/usr/local/bin"
        }
        $TargetBinary = "c8y"
        
        Write-Verbose "Changing execution rights for the binary [$binary]"
        & chmod +x $binary

        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Failed to change binary to executable mode. Try running 'chmod +x $InstallPath/$TargetBinary' manually"
        }
    }
    else {
        if (!$InstallPath) {
            if ($env:HOME) {
                $InstallPath = $env:HOME
            }
        }
        $TargetBinary = "c8y.exe"
    }

    if (!$InstallPath) {
        Write-Warning "InstallPath is empty. Please specify a target install path by using the -InstallPath parameter"
        return
    }

    Write-Verbose "Copying binary to [$InstallPath/$TargetBinary]"

    try {
        $AlreadyInstalled = $false
        if (Test-Path "$InstallPath/$TargetBinary") {
            (Get-FileHash -Path $binary -Algorithm SHA256).Hash -eq (Get-FileHash -Path "$InstallPath/$TargetBinary" -Algorithm SHA256).Hash
        }
        if (!$AlreadyInstalled) {
            Copy-Item -Path $binary -Destination "$InstallPath/$TargetBinary" -ErrorAction Stop
        }
    } catch {
        Write-Warning "Failed to copy file. $_"
        Write-Warning "`nPlease run the following command manually `n`n`tsudo cp `"$binary`" `"$InstallPath/$TargetBinary`""
    }

    if ($env:PATH -notlike "*${InstallPath}*") {
        Write-Warning "The Cumulocity binary has been installed in [$InstallPath] however it is not in your `$PATH variable. This means it will not be accessible from any where in the console. Please add [$InstallPath] to your `$PATH variable"
    }
}
