[cmdletbinding()]
Param()
$RootFolder = $PSScriptRoot

$PublicManual  = @( Get-ChildItem -Path $RootFolder\Public-manual\ -Filter *.ps1 -Recurse -ErrorAction SilentlyContinue )
$Public  = @( Get-ChildItem -Path $RootFolder\Public\ -Filter *.ps1 -Recurse -ErrorAction SilentlyContinue )
$Private = @( Get-ChildItem -Path $RootFolder\Private\ -Filter *.ps1 -Recurse -ErrorAction SilentlyContinue )
$script:Dependencies = & { Join-Path -Path $PSScriptRoot -ChildPath "Dependencies" }


Foreach($import in @($PublicManual + $Public + $Private))
{
    Try
    {
        Write-Verbose ("Importing: {0}" -f $import.FullName)
        . $import.FullName
    }
    Catch
    {
        Write-Error -Message "Failed to import function $($import.fullname): $_"
    }
}

foreach($publicFile in @($PublicManual + $Public)) {
    Write-Verbose "Making: $($publicFile.Basename) public"
    Export-ModuleMember -Function $publicFile.Basename
}

#
# Create session folder
#
$HomePath = Get-SessionHomePath

if (!(Test-Path $HomePath)) {
    Write-Host "Creating home directory [$HomePath]"
    $null = New-Item -Path $HomePath -ItemType Directory
}

# Install binary (and make it executable)
if ($IsLinux -or $IsMacOS) {
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

# Setup encryption (if not already set)
Set-ClientPassphrase

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
    base64ToUtf8 = "ConvertFrom-Base64ToUtf8"

    # session
    session = "Get-Session"
}

Register-Alias

Export-ModuleMember -Alias *
