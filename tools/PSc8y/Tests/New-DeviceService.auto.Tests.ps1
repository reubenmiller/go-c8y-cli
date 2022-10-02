. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceService" {
    BeforeEach {

    }

    It -Skip "Create a child addition and link it to an existing managed object" {
        $Response = PSc8y\New-DeviceService -Id $software.id -Data "custom.value=test" -Global -ChildType addition
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

