. $PSScriptRoot/imports.ps1

Describe -Name "Copy-Application" {
    BeforeEach {
        $App = New-TestHostedApplication

    }

    It "Copy an existing application" {
        $Response = PSc8y\Copy-Application -Id $App.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $App.id
        Remove-Application -Id "clone$($App.name)"

    }
}

