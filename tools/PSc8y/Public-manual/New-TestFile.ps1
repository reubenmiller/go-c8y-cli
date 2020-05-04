Function New-TestFile {
<#
.SYNOPSIS
Create a new temp file with default contents
#>
    Param(
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [object]
        $InputObject = "example message"
    )

    $TempFile = New-TemporaryFile
    $InputObject | Out-File -LiteralPath $TempFile.FullName -Encoding utf8

    $TempFile
}
