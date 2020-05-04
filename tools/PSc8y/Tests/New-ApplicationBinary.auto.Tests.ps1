. $PSScriptRoot/imports.ps1

Describe -Name "New-ApplicationBinary" {
    BeforeEach {
        $AppName = New-RandomString -Prefix "testapp_"
        $App = New-Application -Name $AppName -Type HOSTED -Key "${AppName}-key" -ContextPath $AppName
        $MicroserviceZip = "$PSScriptRoot/TestData/microservice/helloworld.zip"

    }

    It -Skip "Upload application microservice binary" {
        $Response = PSc8y\New-ApplicationBinary -Id $App.id -File $MicroserviceZip
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $App.id

    }
}

