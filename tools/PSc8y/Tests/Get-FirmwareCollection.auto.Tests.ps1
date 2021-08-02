. $PSScriptRoot/imports.ps1

Describe -Name "Get-FirmwareCollection" {
    BeforeEach {

    }

    It "Get a list of firmware packages" {
        $Response = PSc8y\Get-FirmwareCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

