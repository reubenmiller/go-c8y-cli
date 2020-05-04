. $PSScriptRoot/imports.ps1

Describe -Name "Get-ApplicationBinaryCollection" {
    BeforeEach {
        $app = New-TestHostedApplication

    }

    It -Skip "List all of the binaries related to a Hosted (web) application" {
        $Response = PSc8y\Get-ApplicationBinaryCollection -Id $App.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "List all of the binaries related to a Hosted (web) application (using pipeline)" {
        $Response = PSc8y\Get-Application $App.id | Get-ApplicationBinaryCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $app.id

    }
}

