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
    goreleaser build --output "$OutputDir/" --skip-validate --rm-dist
} else {
    # Build for the current environment
    goreleaser build --output "$OutputDir/c8y.linux" --skip-validate --rm-dist --single-target --id linux
}
