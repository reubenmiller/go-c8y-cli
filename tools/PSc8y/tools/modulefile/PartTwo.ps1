
#region imports

# Download c8y binary from github (instead of packing it with the module)
# This allows the module to be used on multiple platforms and CPU architectures without
# increasing the module size
$Platform = ""
$Arch = ""
$BinaryName = ""
$FileExtension = ""
$Version = "2.9.0"

if ($IsLinux) {
    $Arch = switch -regex (uname -m)
    {
        "^armv8l|armv8b|arm64|aarch64$" { "arm64" }
        "^arm$" { "armv5" }
        "^x86_64$" { "amd64" }
        "^i386$" { "386" }
    }

    $Platform = "linux"
    $BinaryName = "c8y.linux"
    $FileExtension = "tar.gz"
} elseif ($IsMacOS) {
    $Arch = switch -regex (uname -m)
    {
        "^armv8l|armv8b|arm64|aarch64$" { "arm64" }
        "^x86_64$" { "amd64" }
        "^i386$" { "386" }
    }

    $Platform = "macOS"
    $BinaryName = "c8y.macos"
    $FileExtension = "tar.gz"
} else {
    # Windows
    $Arch = switch ($Env:PROCESSOR_ARCHITECTURE)
    {
        "amd64" {"amd64"}
        "x86" {"386"}
        "x64" {"amd64"}
        # "arm64" {"arm64"}
    }
    $Platform = "windows"
    $BinaryName = "c8y.windows.exe"
    $FileExtension = "zip"
}

$TargetFile = "c8y_${Version}_${Platform}_${Arch}.${FileExtension}"
$LocalBinary = Join-Path $script:Dependencies $BinaryName
$script:C8Y_BINARY = $LocalBinary
$DownloadUrl = "https://github.com/reubenmiller/go-c8y-cli/releases/download/v$Version/$TargetFile"

if (-Not (Test-Path $script:Dependencies))
{
    $null = New-Item $script:Dependencies -ItemType Directory
}

try
{
    if (-Not (Test-Path $LocalBinary))
    {
        $TempFile = New-TemporaryFile
        Invoke-RestMethod -Uri $DownloadUrl -OutFile $TempFile

        Write-Verbose "Downloading binary from $DownloadUrl to $TempFile"

        switch ($FileExtension)
        {
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

    if ($IsMacOS -or $IsLinux)
    {
        chmod +x $LocalBinary
    }
}
catch
{
    Write-error "Failed to download c8y binary. url=$DownloadUrl, error=$_"
}

# Add alias to c8y binary
Set-Alias -Name "c8y" -Value (Get-ClientBinary) -Scope "Global"

#
# Create session folder
#
$HomePath = Get-SessionHomePath

if ($HomePath -and !(Test-Path $HomePath)) {
    Write-Host "Creating home directory [$HomePath]"
    $null = New-Item -Path $HomePath -ItemType Directory
}

# Install binary (and make it executable)
if ($script:IsLinux -or $script:IsMacOS) {
    # silence errors
    if ($env:PSC8Y_INSTALL_ON_IMPORT -match "true|1|on") {
        Install-ClientBinary -ErrorAction SilentlyContinue
    } else {
        # Make c8y executable
        $binary = Get-ClientBinary
        & chmod +x $binary
    }
}

# Load c8y completions for powershell
c8y completion powershell | Out-String | Invoke-Expression

# Session
Register-ArgumentCompleter -CommandName "Set-Session" -ParameterName Session -ScriptBlock $script:CompletionSession

# Set environment variables if a session is set via the C8Y_SESSION env variable
$ExistingSession = Get-Session -WarningAction SilentlyContinue -ErrorAction SilentlyContinue 2> $null
if ($ExistingSession) {

    # Display current session
    $ConsoleMessage = $ExistingSession | Out-String
    $ConsoleMessage = $ConsoleMessage.TrimEnd()
    Write-Host "Current Cumulocity session"
    Write-Host "${ConsoleMessage}`n"
}

# Enforce UTF8 encoding
$CurrentEncodingName = [Console]::Out.Encoding.EncodingName
$RequiredEncodingName = [System.Text.Encoding]::UTF8.EncodingName

if ($CurrentEncodingName -ne $RequiredEncodingName) {
    if (-not $env:C8Y_DISABLE_ENFORCE_ENCODING) {
        Write-Warning ("Current console encoding is not correct. Changing from [{0}] to [{1}]" -f @(
            $CurrentEncodingName,
            $RequiredEncodingName
        ))
        if ($PROFILE) {
            Write-Warning "You can get rid of this message by adding '[Console]::OutputEncoding = [System.Text.Encoding]::UTF8' to your PowerShell profile ($PROFILE)"
        }
        [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    } else {
        Write-Verbose "User chose to use non-utf8 encoding (via setting C8Y_DISABLE_ENFORCE_ENCODING env variable). Current console encoding is [$CurrentEncodingName] but PSc8y wants [$RequiredEncodingName]"
    }
}

$script:Aliases = @{
    # collections
    alarms = "Get-AlarmCollection"
    apps = "Get-ApplicationCollection"
    devices = "Get-DeviceCollection"
    events = "Get-EventCollection"
    fmo = "Find-ManagedObjectCollection"
    measurements = "Get-MeasurementCollection"
    ops = "Get-OperationCollection"
    series = "Get-MeasurementSeries"

    # single items
    alarm = "Get-Alarm"
    app = "Get-Application"
    event = "Get-Event"
    m = "Get-Measurement"
    mo = "Get-ManagedObject"
    op = "Get-Operation"

    # References
    childdevices = "Get-ChildDeviceCollection"
    childassets = "Get-ChildAssetCollection"

    # utilities
    json = "ConvertTo-NestedJson"
    tojson = "ConvertTo-NestedJson"
    fromjson = "ConvertFrom-JsonStream"
    rest = "Invoke-ClientRequest"
    base64ToUtf8 = "ConvertFrom-Base64String"
    utf8Tobase64 = "ConvertTo-Base64String"
    iterate = "Invoke-ClientIterator"
    batch = "Group-ClientRequests"

    # session
    session = "Get-Session"
}

Register-Alias
#endregion imports
