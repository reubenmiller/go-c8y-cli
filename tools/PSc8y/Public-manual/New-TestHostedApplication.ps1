Function New-TestHostedApplication {
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        [string] $Name,

        [string] $File = "$PSScriptRoot/../Tests/TestData/hosted-application/simple-helloworld",

        [switch] $Force
    )

    if (!$Name) {
        $Name = New-RandomString -Prefix "web-"
    }

    $options = @{
        Name = $Name
    }
    if ($File) {
        $options.File = $File
    }

    New-HostedApplication @options
}
