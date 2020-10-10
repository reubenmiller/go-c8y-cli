Function New-TestMicroservice {
<# 
.SYNOPSIS
Create a new test microservice
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Name of the microservice
        [string] $Name,

        # Microservice zip file to upload to Cumulocity
        [string] $File = "$PSScriptRoot/../Tests/TestData/microservice/helloworld.zip",
        
        # Skip the uploading of the microservice binary and only create the microservice placeholder
        [switch] $SkipUpload,
        
        # Skip the subscription for the microservice
        [switch] $SkipSubscription,

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
        $Name = New-RandomString -Prefix "testms-"
    }

    $options = @{
        Name = $Name
        File = $File
        SkipUpload = $SkipUpload
        SkipSubscription = $SkipSubscription
        Template = $Template
        TemplateVars = $TemplateVars
        Force = $Force
    }

    New-Microservice @options
}
