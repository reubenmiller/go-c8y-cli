. $PSScriptRoot/imports.ps1

Describe -Name "Get-SoftwareCollection" {
    BeforeEach {

    }

    It "Get a list of software packages" {
        $Response = PSc8y\Get-SoftwareCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

