. $PSScriptRoot/imports.ps1

Describe -Name "Remove-DeviceGroup" {
    BeforeEach {
        $group = PSc8y\New-TestDeviceGroup

    }

    It "Remove device group by id" {
        $Response = PSc8y\Remove-DeviceGroup -Id $group.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Remove device group by name" {
        $Response = PSc8y\Remove-DeviceGroup -Id $group.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

