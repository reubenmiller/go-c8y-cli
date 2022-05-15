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
    goreleaser build --skip-validate --rm-dist --snapshot
    # goreleaser build --output "$OutputDir/c8y.macos" --skip-validate --rm-dist --snapshot --single-target --id macos
    # goreleaser build --output "$OutputDir/c8y.windows" --skip-validate --rm-dist --snapshot --single-target --id windows
    # goreleaser build --output "$OutputDir/c8y.linux" --skip-validate --rm-dist --snapshot --single-target --id linux
} else {
    # Build for the current environment
    goreleaser build --output "$OutputDir/c8y.linux" --skip-validate --rm-dist --snapshot --single-target --id linux
}
