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
        [string] $File = "$PSScriptRoot/../Tests/TestData/hosted-application/simple-helloworld",

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

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

    if ($Template) {
        $options.Template = $Template
    }

    if ($TemplateVars) {
        $options.TemplateVars = $TemplateVars
    }

    New-HostedApplication @options
}
