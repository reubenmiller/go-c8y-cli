. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceService" {
    BeforeEach {

    }

    It -Skip "Get service status" {
        $Response = PSc8y\Get-DeviceService -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

