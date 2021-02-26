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
    ConvertFrom-Json @JSONArgs
}
