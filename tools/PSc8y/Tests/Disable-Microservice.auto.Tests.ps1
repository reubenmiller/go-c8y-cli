. $PSScriptRoot/imports.ps1

Describe -Name "Disable-Microservice" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload

    }

    It -Skip "Disable (unsubscribe) to a microservice" {
        $Response = PSc8y\Disable-Microservice -Id $App.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-Microservice -Id $App.id

    }
}

