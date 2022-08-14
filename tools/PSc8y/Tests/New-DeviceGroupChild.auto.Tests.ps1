. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceGroupChild" {
    BeforeEach {

    }

    It -Skip "Create a child addition and link it to an existing managed object" {
        $Response = PSc8y\New-DeviceGroupChild -Id $software.id -Data "custom.value=test" -Global -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

