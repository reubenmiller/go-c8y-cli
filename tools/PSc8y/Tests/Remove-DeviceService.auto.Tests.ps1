. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceService" {
    BeforeEach {

    }

    It -Skip "Remove service" {
        $Response = PSc8y\Remove-DeviceService -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

