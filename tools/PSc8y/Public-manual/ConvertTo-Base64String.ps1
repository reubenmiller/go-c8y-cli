Function ConvertTo-Base64String {
<#
.SYNOPSIS
Convert a UTF8 string to a base64 encoded string

.DESCRIPTION
Convert a UTF8 string to a base64 encoded string

.PARAMETER InputObject
UTF8 encoded string

.EXAMPLE
ConvertTo-Base64String tenant/username:password

Encode a string to base64 encoded string
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
            [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Item))
        }
    }
}
