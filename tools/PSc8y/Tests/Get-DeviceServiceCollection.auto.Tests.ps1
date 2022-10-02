. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceServiceCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of the child additions of an existing managed object" {
        $Response = PSc8y\Get-DeviceServiceCollection -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get a list of the child additions of an existing managed object (using pipeline)" {
        $Response = PSc8y\Get-Device -Id 12345 | Get-DeviceServiceCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

