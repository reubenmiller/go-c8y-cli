. $PSScriptRoot/imports.ps1

Describe -Name "Get-Application" {
    BeforeEach {
        $App = New-TestHostedApplication

    }

    It "Get an application by id" {
        $Response = PSc8y\Get-Application -Id $App.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get an application by name" {
        $Response = PSc8y\Get-Application -Id $App.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $App.id

    }
}

