. $PSScriptRoot/imports.ps1

Describe -Name "Update-DeviceService" {
    BeforeEach {

    }

    It -Skip "Update service status" {
        $Response = PSc8y\Update-DeviceService -Id 12345 -Status up
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

