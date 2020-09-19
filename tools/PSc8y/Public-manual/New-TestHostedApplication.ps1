Function New-TestHostedApplication {
<# 
.SYNOPSIS
Create a new test hosted application

.DESCRIPTION
This is used for testing only

.EXAMPLE
New-TestHostedApplication

Create a test hosted web application
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Hosted application name
        [string] $Name,

        # Application folder or file
        [string] $File = "$PSScriptRoot/../Tests/TestData/hosted-application/simple-helloworld",

        # Don't prompt for confirmation
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
