Function ConvertFrom-Base64ToUtf8 {
<#
.SYNOPSIS
Convert a base64 encoded string to UTF8

.DESCRIPTION
Convert a base64 encoded string to UTF8

.NOTES
If the the string has spaces in it, then only the last part of the string (with no spaces in it) will be used. This makes it easier when trying decode the basic auth string

.PARAMETER InputObject
Base64 encoded string

.EXAMPLE
ConvertFrom-Base64ToUtf8 ZWFzdGVyZWdn

Convert the base64 to utf8

.EXAMPLE
ConvertFrom-Base64ToUtf8 "Authorization: Basic s7sd81kkzyzldjkzkhejhug3kh"

Convert the base64 to utf8
#>
    [CmdletBinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position=0)]
        [string[]] $InputObject
    )

    Process {
        foreach ($Item in $InputObject) {
            $Base64 = ($Item -split "\s+") | Select-Object -Last 1
            [System.Text.Encoding]::UTF8.Getstring([System.Convert]::FromBase64String($Base64))
        }
    }
}
