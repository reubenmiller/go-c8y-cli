. $PSScriptRoot/imports.ps1

Describe -Name "Update-Application" {
    BeforeEach {
        $App = New-TestHostedApplication

    }

    It "Update application availability to MARKET" {
        $Response = PSc8y\Update-Application -Id $App.name -Availability "MARKET"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Application -Id $App.id

    }
}

