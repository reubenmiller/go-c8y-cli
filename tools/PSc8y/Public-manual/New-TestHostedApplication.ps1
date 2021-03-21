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
    [cmdletbinding()]
    Param(
        # Hosted application name
        [string] $Name,

        # Application folder or file
        [string] $File = "$PSScriptRoot/../Tests/TestData/hosted-application/simple-helloworld"
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }
    Process {

        $options = @{} + $PSBoundParameters
        if (-Not $options["Name"]) {
            $options["Name"] = New-RandomString -Prefix "web-"
        }
        New-HostedApplication @options
    }
}
