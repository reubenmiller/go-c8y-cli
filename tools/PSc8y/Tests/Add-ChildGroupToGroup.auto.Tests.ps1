. $PSScriptRoot/imports.ps1

Describe -Name "Add-ChildGroupToGroup" {
    BeforeEach {
        $Group = PSc8y\New-TestDeviceGroup
        $ChildGroup1 = PSc8y\New-TestDeviceGroup
        $CustomGroup = PSc8y\New-TestDeviceGroup
        $SubGroup1 = PSc8y\New-TestDeviceGroup -Type SubGroup
        $SubGroup2 = PSc8y\New-TestDeviceGroup -Type SubGroup

    }

    It "Add a group to a group as a child" {
        $Response = PSc8y\Add-ChildGroupToGroup -Group $Group.id -NewChildGroup $ChildGroup1.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Add multiple devices to a group. Alternatively `Get-DeviceCollection` can be used
to filter for a collection of devices and assign the results to a single group.
" {
        $Response = PSc8y\Get-DeviceGroup $SubGroup1.name, $SubGroup2.name | Add-ChildGroupToGroup -Group $CustomGroup.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $ChildGroup1.id
        PSc8y\Remove-ManagedObject -Id $Group.id
        PSc8y\Remove-ManagedObject -Id $SubGroup1.id
        PSc8y\Remove-ManagedObject -Id $SubGroup2.id
        PSc8y\Remove-ManagedObject -Id $CustomGroup.id

    }
}

