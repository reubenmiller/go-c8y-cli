
#region imports

# Download c8y binary from github (instead of packing it with the module)
# This allows the module to be used on multiple platforms and CPU architectures without
# increasing the module size
Import-ClientBinary

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
