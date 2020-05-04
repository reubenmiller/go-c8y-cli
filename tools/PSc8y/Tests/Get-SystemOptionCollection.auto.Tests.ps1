. $PSScriptRoot/imports.ps1

Describe -Name "Get-SystemOptionCollection" {
    BeforeEach {

    }

    It "Get a list of system options" {
        $Response = PSc8y\Get-SystemOptionCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

