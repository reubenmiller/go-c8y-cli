#!/usr/bin/env pwsh
[cmdletbinding()]
Param(
    # installation path
    [string] $InstallPath = "~/bin",

    # User home folder, where the go-c8y-cli-addons will be installed
    [string] $UserHome = "~",

    # go-c8y-cli version to install
    [string] $Version = "latest",

    # Path to profile which should be edited to add the custom location to it
    [string] $ProfilePath = $PROFILE,

    # Skip the version check even if c8y is already installed
    [switch] $SkipVersionCheck
)
<# 
.EXAMPLE
./install.ps1

Install go-c8y-cli in the default location. Only install if the current version does not match the latest

.EXAMPLE
./install.ps1 -SkipVersionCheck

Install go-c8y-cli in the default location and ignore any version checking (even if you already have the latest version, it will still be downloaded again)

.EXAMPLE
./install.ps1 -InstallPath /opt/tools/cumulocity -UserHome /opt/tools/cumulocity -SkipVersionCheck

Install go-c8y-cli and addons to a shared location accessible by multiple users

.EXAMPLE
./install.ps1 -InstallPath C:\tools\cumulocity -UserHome C:\tools\cumulocity -SkipVersionCheck

Force re-installation of go-c8y-cli and addons to a shared location. go-c8y-cli will be installed again even the binary already exists.

.EXAMPLE
powershell -executionPolicy bypass C:\Users\myuser\.go-c8y-cli\install.ps1 -InstallPath C:\tools\cumulocity -UserHome C:\tools\cumulocity -SkipVersionCheck -ProfilePath $PROFILE.AllUsersAllHosts

Install/re-install go-c8y-cli and edit the global powershell profile for all users. This needs to be run as admin (and use a full path to the install.ps1 script)
#>

$UserHome = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($UserHome)

# Expand install path (but it might not yet exist)
$InstallPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($InstallPath)

$OWNER = "reubenmiller"
$REPO = "go-c8y-cli"
$AddonRepo = "go-c8y-cli-addons"

# Function Invoke-CheckoutAddons {
#     [cmdletbinding()]
#     Param()

#     $Destination = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath("$UserHome/.go-c8y-cli")

#     if (Test-Path $Destination) {
#         Write-Verbose "Addon repository has already been cloned to $Destination"
#         return
#     }

#     if (Get-Command git -ErrorAction SilentlyContinue) {
#         Write-Host "Cloning $AddonRepo repository"
#         git clone "https://github.com/$OWNER/${AddonRepo}.git" $Destination
#     } else {
#         # TODO: Manually download addons
#         # Invoke-WebRequest "https://github.com/reubenmiller/go-c8y-cli-addons/archive/refs/heads/main.zip"
#         Write-Warning "Could not find git. Please install git if you want to be able to install the addons"
#     }
# }

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

        "v" + "$current_version".Trim()
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
        [string] $Tag
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
        # Uri = "https://api.github.com/repos/$OWNER/$REPO/releases/assets/$assetId"
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

Function install-c8ybinary () {
    [cmdletbinding()]
    Param(
        [string] $Version,

        [switch] $SkipVersionCheck
    )

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

    Invoke-DownloadC8yBinary -BaseURL $ReleaseBaseURL -Version $Version -Tag $Tag

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

Function Add-ToProfile {
    [cmdletbinding()]
    Param()
    $ProfileDir = Split-Path $ProfilePath -Parent
    if ($ProfileDir -and -Not (Test-Path $ProfileDir)) {
        Write-Verbose "Creating powershell profile directory"
        $null = New-Item -Path $ProfileDir -ItemType Directory
    }
    
    if (-Not (Test-Path $ProfilePath)) {
        "" | Out-File -FilePath $ProfilePath
    }

    $PathStatement = if ($IsMacOS -or $IsLinux) {
        "`$env:PATH = `"${InstallPath}:`$env:PATH`""
    } else {
        "`$env:PATH = `"${InstallPath};`$env:PATH`""
    }

    $CustomSessionHome = Join-Path $UserHome -ChildPath ".cumulocity"
    if (Test-Path $CustomSessionHome) {
        if (-Not (Select-String -Path $ProfilePath -Pattern "C8Y_SESSION_HOME" -Quiet)) {
            Write-Verbose "Adding detected .cumulocity session home to profile"
            Add-Content -Path $ProfilePath -Value "`$env:C8Y_SESSION_HOME = '$CustomSessionHome'`n"
            # Also set it now so it is available now
            $env:C8Y_SESSION_HOME = "$CustomSessionHome"
        }
    }
    
    $ImportSnippet = @(
        $PathStatement,
        ". `"$UserHome/.go-c8y-cli/shell/c8y.plugin.ps1`""
    )
    if (-Not (Select-String -Path $ProfilePath -SimpleMatch -Pattern "env:C8Y_HOME" -Quiet)) {
        $CustomSettings = @(
            "`$env:C8Y_HOME = `"$UserHome/.go-c8y-cli`"",
            "`$env:C8Y_SESSION_HOME = `"$UserHome/.cumulocity`""
        )
        Write-Host "Adding custom settings to profile: $ProfilePath"
        Add-Content -Path $ProfilePath -Value (($CustomSettings -join "`n") + "`n")
    }

    if (-Not (Select-String -Path $ProfilePath -SimpleMatch -Pattern "$UserHome/.go-c8y-cli/shell/c8y.plugin.ps1" -Quiet)) {
        Write-Host "Adding imports to profile: $ProfilePath"
        Add-Content -Path $ProfilePath -Value (($ImportSnippet -join "`n") + "`n")
    }

    # Importing script (for immediate usage)
    . "$UserHome/.go-c8y-cli/shell/c8y.plugin.ps1"
}

#
# Main
#
# Invoke-CheckoutAddons
install-c8ybinary -Version $Version -SkipVersionCheck:$SkipVersionCheck
Add-ToProfile
