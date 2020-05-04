. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Microservice" {
    BeforeEach {
        $App = New-TestMicroservice -SkipUpload -SkipSubscription

    }

    It -Skip "Delete a microservice by id" {
        $Response = PSc8y\Remove-Microservice -Id $App.id
        $LASTEXITCODE | Should -Be 0
    }

    It -Skip "Delete a microservice by name" {
        $Response = PSc8y\Remove-Microservice -Id $App.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

