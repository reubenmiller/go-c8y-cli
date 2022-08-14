. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceChild" {
    BeforeEach {

    }

    It "Create a child addition and link it to an existing managed object" {
        $Response = PSc8y\New-DeviceChild -Id $software.id -Data "custom.value=test" -Global -ChildType childAdditions
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

