. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceServiceCollection" {
    BeforeEach {

    }

    It -Skip "Get services for a specific device" {
        $Response = PSc8y\Get-DeviceServiceCollection -Device 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get services for a specific device (using pipeline)" {
        $Response = PSc8y\Get-Device -Id 12345 | Get-DeviceServiceCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

