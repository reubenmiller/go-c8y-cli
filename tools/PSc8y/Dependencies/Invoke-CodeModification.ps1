[cmdletbinding(
    SupportsShouldProcess = $true,
    ConfirmImpact = "High"
)]
Param(
    # Path to the ps1 files which should be modified
    [Parameter(
        Mandatory = $true,
        Position = 0)]
    [string] $Path,

    [switch] $Force
)

$commands = Get-Command -Module PSc8y | Select-Object -ExpandProperty Name

$SourcesFiles = Get-ChildItem -Path $Path -Filter "*.ps1" -Recurse

# Modifications
$Mods = @(
    @{
        name = "Replace -ExternalId parameter"
        pattern = " -ExternalId\b"
        replacement = " -Name"
    },

    @{
        name = "Remove -SkipQueryParser parameter"
        pattern = " ?-SkipQueryParser(:\`$(true|false))?\b"
        replacement = ""
    },

    @{
        name = "Renamed Find-ManagedObject"
        pattern = "Find-(C8y)?ManagedObject\b"
        replacement = "Find-ManagedObjectCollection"
    },

    @{
        name = "Renamed Get-Identity"
        pattern = "Get-(C8y)?Identity\b"
        replacement = "Get-ExternalId"
    },

    @{
        name = "Get-Binary"
        # Within Get-C8yBInary, rename "-OutFile" to "-OutputFile"
        pattern = "Get-(C8y)?Binary\b.*-OutFile"
        replacement = "-OutputFile"
    },

    @{
        name = "Update-ManagedObject"
        # Within Update-C8yManagedObject, rename "-Property" to "-Data"
        pattern = "Update-(C8y)?ManagedObject\b.*-Property"
        replacement = "-Data"
    }
)

foreach ($iFile in $SourcesFiles) {
    $original = Get-Content $iFile.Fullname -Raw
    $contents = $original

     # Apply modifications
    foreach ($iMod in $Mods) {
        $contents = $contents -replace $iMod.pattern, $iMod.replacement
    }

    # Force use of fully qualified command names
    foreach ($iCommand in $commands) {
        $iCommandPattern = $iCommand -replace "-", "-(C8y)?"
        # Note: Don't add a prefix behind the
        $contents = $contents -ireplace "(?<!Function )(PSc8y\\)?$iCommandPattern\b", "PSc8y\$iCommand"
    }

    if ($contents -ne $original) {
        Write-Verbose ("Updating file {0}" -f $iFile.FullName)

        if (!$Force -and !$PSCmdlet.ShouldProcess("Apply code patch to", $iFile.FullName)) {
            continue
        }
        $contents | Out-File -FilePath $iFile.FullName -Encoding utf8
    }
}
