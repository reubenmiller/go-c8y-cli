. $PSScriptRoot/imports.ps1

Describe -Name "New-MicroserviceBinary" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload -SkipSubscription
        $MicroserviceZip = "$PSScriptRoot/TestData/microservice/helloworld.zip"

    }

    It -Skip "Upload microservice binary" {
        $Response = PSc8y\New-MicroserviceBinary -Id $App.id -File $MicroserviceZip
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Microservice -Id $App.id

    }
}

