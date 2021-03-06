﻿Function New-TestFile {
<#
.SYNOPSIS
Create a new temp file with default contents

.DESCRIPTION
Create a temporary file with some contents which can be used to uploaded it to Cumulocity
via the Binary api.

.EXAMPLE
New-TestFile

Create a temp file with pre-defined content

.EXAMPLE
"My custom text info" | New-TestFile

Create a temp file with customized content.

.OUTPUTS
System.IO.FileInfo

#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "Low"
    )]
    Param(
        # Content which should be written to the temporary file
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [object]
        $InputObject = "example message",

        # Don't prompt for confirmation
        [switch]
        $Force
    )

    if (!$Force -and
        !$WhatIfPreference -and
        !$PSCmdlet.ShouldProcess("Create a temporary file")) {
        return
    }

    $TempFile = New-TemporaryFile
    $InputObject | Out-File -LiteralPath $TempFile.FullName -Encoding utf8

    $TempFile
}
