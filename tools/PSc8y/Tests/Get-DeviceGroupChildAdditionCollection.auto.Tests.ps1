. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupChildAdditionCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of the child additions of an existing device" {
        $Response = PSc8y\Get-DeviceGroupChildAdditionCollection -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

