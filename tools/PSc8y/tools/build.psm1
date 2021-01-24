
$script:ModuleName = "PSc8y"
$script:ProjectRoot = Split-Path -Path $PSScriptRoot -Parent
$script:ModuleRoot = $ProjectRoot
$script:ArtifactRoot = Join-Path -Path $ProjectRoot -ChildPath "dist"
$script:TempPath = [System.IO.Path]::GetTempPath()

$script:PublicFunctions = @( Get-ChildItem -Path $ModuleRoot\Public\*.ps1 -ErrorAction SilentlyContinue ) +
                          @(
    Get-ChildItem -Path $ModuleRoot\Public-manual\*.ps1 -ErrorAction SilentlyContinue `
        | Where-Object { $_.Name -ne "New-TestMicroservice.ps1" -and $_.Name -ne "New-TestHostedApplication.ps1" }
)
$script:PrivateFunctions = @( Get-ChildItem -Path $ModuleRoot\Private\*.ps1 -ErrorAction SilentlyContinue )

$script:EnumDefinitions = @( Get-ChildItem -Path "$ModuleRoot\Enums\*.ps1" -ErrorAction SilentlyContinue )

function New-ModulePSMFile {
    $moduleFile = New-Item -Path $ArtifactRoot\$ModuleName\$ModuleName.psm1 -ItemType File -Force

    # Add the first part of the distributed .psm1 file from template.
    Get-Content -Path "$ModuleRoot\tools\modulefile\PartOne.ps1" | Out-File -FilePath $moduleFile -Append -Encoding utf8BOM

    # Add a region and write out the private functions.
    "`n#region Private Functions" | Out-File -FilePath $moduleFile -Append
    Get-Content $PrivateFunctions | Out-String | Out-File -FilePath $moduleFile -Append
    "#endregion`n" | Out-File -FilePath $moduleFile -Append

    # Add a region and write out the public functions
    "#region Public Functions" | Out-File -FilePath $moduleFile -Append
    Get-Content $PublicFunctions | Out-String | Out-File -FilePath $moduleFile -Append
    "#endregion`n" | Out-File -FilePath $moduleFile -Append

    # Add a region for Enums
    "#region Enums" | Out-File -FilePath $moduleFile -Append
    Get-Content $EnumDefinitions | Out-String | Out-File -FilePath $moduleFile -Append
    "#endregion`n" | Out-File -FilePath $moduleFile -Append

    # Build a string to export only /public/psmexports functions from the PSModule.psm1 file.
    $publicFunctionNames = @( $PublicFunctions.BaseName )
    foreach ($publicFunction in $publicFunctionNames) {
        $functionNameString += "$publicFunction,"
    }

    $functionNameString = $functionNameString.TrimEnd(",")
    $functionNameString = "Export-ModuleMember -Function $functionNameString`n"

    # Add the export module member string to the module file.
    $functionNameString | Out-File -FilePath $moduleFile -Append

    # Add the remaining part of the psm1 file from template.
    Get-Content -Path "$ModuleRoot\tools\modulefile\PartTwo.ps1" | Out-File -FilePath $moduleFile -Append
}

function Update-ModuleManifestFunctions {
    # Update the psd1 file with the /public/psgetfunctions
    # Update-ModuleManifest is not used because a) it is not availabe for ps version <5.0 and b) it is destructive.
    # First a helper method removes the functions and replaces with the standard FunctionsToExport = @()
    # then this string is replaced by another string built from /public/psgetfunctions

    $ManifestFile = "$ModuleRoot\$ModuleName.psd1"

    # Call helper function to replace with an empty FunctionsToExport = @()
    Remove-ModuleManifestFunctions -Path $ManifestFile

    $ManifestFileContent = Get-Content -Path "$ManifestFile"

    # FunctionsToExport string needs to be array definition with function names surrounded by quotes.
    $formatedFunctionNames = @()
    foreach ($function in $PublicFunctions.BaseName) {
        $function = "`'$function`'"
        $formatedFunctionNames += $function
    }

    # Tabbing and new lines to make the psd1 consistent
    $formatedFunctionNames = $formatedFunctionNames -join ",`n`t"
    $ManifestFunctionExportString = "FunctionsToExport = @(`n`t$formatedFunctionNames)`n"

    # Do the string replacement in the manifest file with the formated function names.
    $ManifestFileContent = $ManifestFileContent.Replace('FunctionsToExport = "*"', $ManifestFunctionExportString)
    Set-Content -Path "$ManifestFile" -Value $ManifestFileContent.TrimEnd()
}
function Remove-ModuleManifestFunctions ($Path) {
    # Utility method to remove the list of functions from a manifest. This is specific to this modules manifest and
    # assumes the next item in the manifest file after the functions is a comment containing 'VariablesToExport'.

    $rawFile = Get-Content -Path $Path -Raw
    $arrFile = Get-Content -Path $Path

    $functionsStartPos = ($arrFile | Select-String -Pattern 'FunctionsToExport =').LineNumber - 1
    $functionsEndPos = ($arrFile | Select-String -Pattern 'VariablesToExport =').LineNumber - 2

    $functionsExportString = $arrFile[$functionsStartPos..$functionsEndPos] | Out-String

    $rawFile = $rawFile.Replace($functionsExportString, "FunctionsToExport = `"*`"`n")

    Set-Content -Path $Path -Value $rawFile.TrimEnd()
}

function Publish-ModuleArtifacts {

    if (Test-Path -Path $ArtifactRoot) {
        # Note: Remove-item fails in DevContainer for some unknown reason,
        # so fallback to default
        try {
            Remove-Item -LiteralPath $ArtifactRoot -Recurse -Force -ErrorAction Stop
        } catch {
            Write-Warning "Failed to remove folder using Remove-Item, using rm instead"
            rm -Rf "$ArtifactRoot"
        }
    }

    Write-Verbose "Creating directory [$ArtifactRoot]"
    New-Item -Path $ArtifactRoot -ItemType Directory | Out-Null

    # Copy the module into the dist folder
    Copy-Item -Path "$ModuleRoot\Dependencies\" -Filter "c8y*" -Destination "$ArtifactRoot\$ModuleName\Dependencies" -Recurse
    Copy-Item -Path "$ModuleRoot\format-data" -Destination "$ArtifactRoot\$ModuleName\" -Recurse
    Copy-Item -Path "$ModuleRoot\$ModuleName.psd1" -Destination "$ArtifactRoot\$ModuleName\" -Recurse

    # Construct the distributed .psm1 file.
    New-ModulePSMFile

    # Package the module in /dist
    $zipFileName = "$ModuleName.zip"
    $artifactZipFile = Join-Path -Path $ArtifactRoot -ChildPath $zipFileName
    $tempZipfile = Join-Path -Path $TempPath -ChildPath $zipFileName

    if ($PSEdition -ne 'Core') {
        Add-Type -assemblyname System.IO.Compression.FileSystem
    }

    if (Test-Path -Path $tempZipfile) {
        Remove-Item -Path $tempZipfile -Force
    }

    Write-Verbose "Zipping module artifacts in $ArtifactRoot"
    [System.IO.Compression.ZipFile]::CreateFromDirectory($ArtifactRoot, $tempZipfile)

    Move-Item -Path $tempZipfile -Destination $artifactZipFile -Force
}

function Export-ProductionModule {
    Update-ModuleManifestFunctions
    Publish-ModuleArtifacts

    $ExportPath = Join-Path -Path $ArtifactRoot -ChildPath $ModuleName

    Write-Host "`n    Created module in: ${ExportPath}`n" -ForegroundColor Gray

    $ExportPath
}
