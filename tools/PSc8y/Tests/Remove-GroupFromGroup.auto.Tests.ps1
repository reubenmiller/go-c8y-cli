. $PSScriptRoot/imports.ps1

Describe -Name "Remove-GroupFromGroup" {
    BeforeEach {
        $Group = PSc8y\New-TestDeviceGroup
        $ChildGroup = PSc8y\New-TestDevice
        PSc8y\Add-AssetToGroup -Group $Group.id -NewChildGroup $ChildGroup.id

    }

    It "Unassign a child group from its parent" {
        $Response = PSc8y\Remove-GroupFromGroup -Id $Group.id -Child $ChildGroup.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $ChildGroup.id
        PSc8y\Remove-ManagedObject -Id $Group.id

    }
}

