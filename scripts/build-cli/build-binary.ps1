[cmdletbinding()]
Param(
    [Parameter(
        Mandatory = $true,
        Position = 0)]
    [string] $OutputDir,

    [switch] $CompressOnly,

    # Build binaries for all
    [switch] $All
)

if ($All) {
    $env:BINARY_INCLUDE_VERSION = "true"
    goreleaser build --skip-validate --rm-dist --snapshot

    # Copy commonly used binaries to the output directory
    # Note: This might be removed in the future to make the distribution size of the PowerShell module smaller
    Get-Item .
    Get-ChildItem "dist/linux_linux_amd64_v1" -Recurse

    Write-Host "OutputDir: $OutputDir"
    Get-Item $OutputDir

    Copy-Item "dist/macos_darwin_amd64_v1/bin/c8y*" "$OutputDir/"
    Copy-Item "dist/linux_linux_amd64_v1/bin/c8y*" "$OutputDir/"
    Copy-Item "dist/windows_windows_amd64_v1/bin/c8y*" "$OutputDir/"
    # goreleaser build --output "$OutputDir/c8y.macos" --skip-validate --rm-dist --snapshot --single-target --id macos
    # goreleaser build --output "$OutputDir/c8y.windows" --skip-validate --rm-dist --snapshot --single-target --id windows
    # goreleaser build --output "$OutputDir/c8y.linux" --skip-validate --rm-dist --snapshot --single-target --id linux
} else {
    # Build for the current environment
    goreleaser build --output "$OutputDir/c8y.linux" --skip-validate --rm-dist --snapshot --single-target --id linux
}
