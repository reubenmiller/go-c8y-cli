. $PSScriptRoot/imports.ps1

Describe -Name "Remove-ChildAssetReference" {
    BeforeEach {
        $Group = PSc8y\New-TestDeviceGroup
        $ChildDevice = PSc8y\New-TestDevice
        PSc8y\New-ChildAssetReference -Group $Group.id -NewChildDevice $ChildDevice.id

    }

    It "Unassign a child device from its parent asset" {
        $Response = PSc8y\Remove-ChildAssetReference -Asset $Group.id -ChildDevice $ChildDevice.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $ChildDevice.id
        PSc8y\Remove-ManagedObject -Id $Group.id

    }
}

