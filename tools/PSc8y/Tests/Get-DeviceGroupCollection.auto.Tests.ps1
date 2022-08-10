. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupCollection" {
    BeforeEach {

    }

    It "Get a collection of device groups with names that start with 'MyGroup'" {
        $Response = PSc8y\Get-DeviceGroupCollection -Name "MyGroup*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

