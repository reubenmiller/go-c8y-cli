. $PSScriptRoot/imports.ps1

Describe -Name "Add-DeviceSoftware" {
    BeforeEach {

    }

    It -Skip "Add software to a device" {
        $Response = PSc8y\Add-DeviceSoftware -Device 12345 -Name myapp -Version 1.0.2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

