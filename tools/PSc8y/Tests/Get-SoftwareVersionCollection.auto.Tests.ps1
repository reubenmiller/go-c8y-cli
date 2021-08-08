. $PSScriptRoot/imports.ps1

Describe -Name "Get-SoftwareVersionCollection" {
    BeforeEach {

    }

    It "Get a list of software package versions" {
        $Response = PSc8y\Get-SoftwareVersionCollection -SoftwareId 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

