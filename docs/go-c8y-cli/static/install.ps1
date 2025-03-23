new-module -name c8y-installer -scriptblock {
    $OWNER = "reubenmiller"
    $REPO = "go-c8y-cli"
    $AddonRepo = "go-c8y-cli-addons"


    # TODO: Check for arm64
    Function Get-CPUArchitecture() {
        if ([Environment]::Is64BitOperatingSystem) {
            "amd64"
        } else {
            "386"
        }
    }

    Function Get-OSVersion {
        if ($IsMacOS) {
            "macOS"
        } elseif ($IsLinux) {
            "linux"
        } else {
            "windows"
        }
    }

    Function Get-CurrentVersion {
        if (Get-Command "c8y" -ErrorAction SilentlyContinue) {
            $current_version = & c8y version --select version --output csv 2> $null

            if ([string]::IsNullOrEmpty($current_version)) {
                return
            }

            "$current_version".Trim()
        }
    }

    function New-TemporaryDirectory {
        $parent = [System.IO.Path]::GetTempPath()
        [string] $name = [System.Guid]::NewGuid()
        New-Item -ItemType Directory -Path (Join-Path $parent $name)
    }

    Function Invoke-DownloadC8yBinary {
        [cmdletbinding()]
        Param(
            [string] $BaseURL,
            [string] $Version,
            [string] $Tag,
            [string] $InstallPath
        )

        $os = Get-OSVersion
        $arch = Get-CPUArchitecture

        if ($InstallPath -and -Not (Test-Path $InstallPath)) {
            $null = New-Item -ItemType Directory -Path $InstallPath
        }

        $binaryName = "c8y"
        $Version = $Version -replace "^v", ""
        $package = "c8y_${Version}_${os}_${arch}"
        $archive = "${package}.tar.gz"

        if ($os -eq "windows") {
            $archive = "${package}.zip"
            $binaryName = "c8y.exe"
        }

        $tmp = [system.io.path]::GetTempPath()
        $DownloadedFile = Join-Path -path $tmp -ChildPath $archive
        
        Invoke-DownloadAsset -Tag $Tag -FileName $archive -OutFile $DownloadedFile

        if ($DownloadedFile -match ".zip") {
            $tmp = New-TemporaryDirectory
            Microsoft.PowerShell.Archive\Expand-Archive -Path $DownloadedFile -DestinationPath "$tmp/$package"
            Write-Host "Installing c8y to $InstallPath"
            Copy-Item "$tmp/$package/bin/c8y*" -Destination "$InstallPath/"
        } else {
            if (Get-Command "tar" -ErrorAction SilentlyContinue) {
                tar zxf "$tmp/$archive" -C "$tmp"

                Write-Host "Installing c8y to $InstallPath"
                Copy-Item "$tmp/$package/bin/c8y*" -Destination "$InstallPath/"
            } else {
                Write-Error "Could not find tar and it is required to extract archive"
            }
        }

        
        Remove-Item -Path "$tmp/$package" -Recurse -Force
        
        if ($IsMacOS -or $IsLinux) {
            chmod a+x $InstallPath/$BinaryName
        }
    }

    Function Invoke-DownloadAsset {
        [cmdletbinding()]
        Param(
            [string] $Tag,
            [string] $FileName,
            [string] $OutFile

        )
        $options = @{
            Uri = "https://api.github.com/repos/$OWNER/$REPO/releases/tags/$tag"
            Headers = @{
                Accept = "application/vnd.github.v3+json"
            }
        }

        if (![string]::IsNullOrWhiteSpace($env:CURL_AUTH_HEADER)) {
            $options.Headers["Authorization"] = "Bearer $GITHUB_TOKEN"
        }

        $release_info = Invoke-RestMethod @options

        $asset = $release_info.assets `
        | Where-Object { $_.name -eq $FileName } `
        | Select-Object -First 1
        
        
        if ($null -eq $asset) {
            Write-Warning "Could not find download artifact"
            return
        }
    
        $options = @{
            Uri = $asset.browser_download_url
            Headers = @{
                Accept = "application/octet-stream"
            }
            OutFile = $OutFile
        }
        
        try {
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest @options
        } catch {
        } finally {
            $ProgressPreference = 'Continue'
        }
    }

    Function install-c8y {
        <#
        .SYNOPSIS
        Install go-c8y-cli

        .EXAMPLE
        ./install.ps1

        Install go-c8y-cli in the default location. Only install if the current version does not match the latest

        .EXAMPLE
        ./install.ps1 -SkipVersionCheck

        Install go-c8y-cli in the default location and ignore any version checking (even if you already have the latest version, it will still be downloaded again)


        .EXAMPLE
        ./install.ps1 -InstallPath C:\tools\cumulocity -SkipVersionCheck

        Force re-installation of go-c8y-cli and addons to a shared location. go-c8y-cli will be installed again even the binary already exists.

        #>
        [cmdletbinding()]
        Param(
            [string] $Version = "latest",

            # installation path
            [string] $InstallPath = "~/bin",

            [switch] $SkipVersionCheck
        )

        # Expand install path (but it might not yet exist)
        $InstallPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($InstallPath)

        $ReleaseInfo = Get-LatestTag

        if ($Version -eq "latest") {
            $Version = $ReleaseInfo.Tag
        }
        $Tag = $ReleaseInfo.Tag

        $ReleaseBaseURL = "https://github.com/$OWNER/$REPO/releases/download/$Version"

        if ($ReleaseInfo.BinaryBaseUrl) {
            $ReleaseBaseURL = $ReleaseInfo.BinaryBaseUrl
        }

        if ($Version -ne $Tag) {
            Write-Host "Latest version: $Version (tag=$Tag)"
        } else {
            Write-Host "Latest version: $Tag"
        }

        $CurrentVersion = Get-CurrentVersion

        if (-Not $SkipVersionCheck -and $CurrentVersion -eq $Version) {
            Write-Host "c8y is already up to date: $Version" -ForegroundColor Green
            return
        }

        if ($CurrentVersion) {
            Write-Host "Updating from $CurrentVersion to $Version"    
        } else {
            Write-Host "Installing $Version"
        }

        Invoke-DownloadC8yBinary -BaseURL $ReleaseBaseURL -Version $Version -Tag $Tag -InstallPath $InstallPath

        if ($env:PATH -notlike "*${InstallPath}*") {
            if ($IsLinux -or $IsMacOS) {
                $env:PATH = $InstallPath + ":" + $env:PATH
            } else {
                $env:PATH = $InstallPath + ";" + $env:PATH
            }
        }


        if (-Not (Get-Command "c8y" -ErrorAction SilentlyContinue)) {
            if ($IsLinux -or $IsMacOS) {
                $env:PATH = $InstallPath + ":" + $env:PATH
            } else {
                $env:PATH = $InstallPath + ";" + $env:PATH
            }
        }

        # install profiles
        & $InstallPath/c8y cli install

        # show new version
        & $InstallPath/c8y version
    }

    Function Get-LatestTag () {
        [cmdletbinding()]
        Param()
        $options = @{
            Uri = "https://api.github.com/repos/$OWNER/$REPO/releases"
            Headers = @{
                Accept = "application/vnd.github.v3+json"
            }
        }
        if (![string]::IsNullOrWhiteSpace($env:CURL_AUTH_HEADER)) {
            $options.Headers["Authorization"] = "Bearer $GITHUB_TOKEN"
        }

        $resp = Invoke-RestMethod @options
        $TagName = $resp[0].tag_name
        $BinaryName = Split-Path $resp[0].assets[0].browser_download_url -Leaf
        $BinaryBaseUrl = $resp[0].assets[0].browser_download_url -replace "\/[^\/]+$", ""

        New-Object pscustomobject -Property @{
            Tag = $TagName
            BinaryName = $BinaryName
            BinaryBaseUrl = $BinaryBaseUrl
        }
    }
    Export-ModuleMember -Function install-c8y
}
