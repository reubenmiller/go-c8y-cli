
#region imports
#
# Create session folder
#
$HomePath = Get-SessionHomePath

if (!(Test-Path $HomePath)) {
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

# Set environment variables if a session is set via the C8Y_SESSION env variable
$ExistingSession = Get-Session -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
if ($ExistingSession) {
    Set-EnvironmentVariablesFromSession

    # Display current session
    $ConsoleMessage = $ExistingSession | Out-String
    $ConsoleMessage = $ConsoleMessage.TrimEnd()
    Write-Host "Current Cumulocity session"
    Write-Host "${ConsoleMessage}`n"
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
    json = "ConvertTo-Json"
    tojson = "ConvertTo-Json"
    fromjson = "ConvertFrom-Json"
    rest = "Invoke-ClientRequest"
    base64ToUtf8 = "ConvertFrom-Base64String"
    utf8Tobase64 = "ConvertTo-Base64String"

    # session
    session = "Get-Session"
}

Register-Alias
#endregion imports

#region tab completion
# allow -Session params to be tab-completed

if (Get-Command -Name "Import-PowerShellDataFile" -ErrorAction SilentlyContinue) {
    # Note: Test-ModuleManifest sometimes throws an error:
    # "collection was modified; enumeration operation may not execute"
    # Import-PowerShellDataFile seems to be more reliable
    $Manifest = Import-PowerShellDataFile -Path $PSScriptRoot\PSc8y.psd1
} else {
    $Manifest = Test-ModuleManifest -Path $PSScriptRoot\PSc8y.psd1
}

$ModulePrefix = $Manifest.Prefix

$ModuleCommands = @( $Manifest.ExportedFunctions.Keys ) `
    | ForEach-Object {
        # Note: Different PowerShell version handle internal function names 
        # slightly differenty (some with prefix sometimes without), so we always
        # look for both of them.
        #
        $Name = "$_"
        $NameWithoutPrefix = $Name.Replace("-${ModulePrefix}", "-")

        if (Test-Path "Function:\$Name") {
            Get-Item "Function:\$Name"
        } elseif (Test-Path "Function:\$NameWithoutPrefix") {
            Get-Item "Function:\$NameWithoutPrefix"
        } else {
            throw "Could not find function '$Name'"
        }
    }

try {
    if (Get-Command -Name Register-ArgumentCompleter -ErrorAction SilentlyContinue) {
        $ModuleCommands | Register-ClientArgumentCompleter
    }
}
catch {
    # All this functionality is optional, so suppress errors
    Write-Debug -Message "Error registering argument completer: $_"
}

#endregion tab completion
