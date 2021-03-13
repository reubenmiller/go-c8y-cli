Function New-TestMicroservice {
<# 
.SYNOPSIS
Create a new test microservice
#>
    [cmdletbinding()]
    Param(
        # Name of the microservice
        [string] $Name,

        # Microservice zip file to upload to Cumulocity
        [string] $File = "$PSScriptRoot/../Tests/TestData/microservice/helloworld.zip",
        
        # Skip the uploading of the microservice binary and only create the microservice placeholder
        [switch] $SkipUpload,
        
        # Skip the subscription for the microservice
        [switch] $SkipSubscription,

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
        Force = $Force
    }

    New-Microservice @options
}
