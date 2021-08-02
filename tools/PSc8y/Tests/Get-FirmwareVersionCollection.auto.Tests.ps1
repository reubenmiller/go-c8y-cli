. $PSScriptRoot/imports.ps1

Describe -Name "Get-FirmwareVersionCollection" {
    BeforeEach {

    }

    It "Get a list of firmware package versions" {
        $Response = PSc8y\Get-FirmwareVersionCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

