Function New-TestMicroservice {
    [cmdletbinding()]
    Param(
        [string] $Name,

        [string] $File = "$PSScriptRoot/../Tests/TestData/microservice/helloworld.zip",
        
        [switch] $SkipUpload,
        
        [switch] $SkipSubscription,

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
