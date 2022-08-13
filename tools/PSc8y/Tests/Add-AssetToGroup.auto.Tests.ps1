. $PSScriptRoot/imports.ps1

Describe -Name "Add-AssetToGroup" {
    BeforeEach {
        $Group1 = PSc8y\New-TestDeviceGroup
        $Group2 = PSc8y\New-TestDeviceGroup

    }

    It "Create group hierarchy (parent group -> child group)" {
        $Response = PSc8y\Add-AssetToGroup -Group $Group1.id -ChildGroup $Group2.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Group2.id
        PSc8y\Remove-ManagedObject -Id $Group1.id

    }
}

