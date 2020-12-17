Function Get-JSONFromResponse {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Response
    )

    $BodyText = "{}"
    if ($Response -match "(?ms)Body:\s*(\{.*\})") {
        $BodyText = $Matches[1]
    }

    $JSONArgs = @{
        InputObject = $BodyText
    }
    if ($PSVersionTable.PSVersion.Major -gt 5) {
        $JSONArgs.Depth = 100
    }
    ConvertFrom-Json @JSONArgs
}
