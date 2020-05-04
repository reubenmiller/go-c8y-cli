. $PSScriptRoot/imports.ps1

Describe -Name "New-ChildAssetReference" {
    BeforeEach {
        $Group1 = PSc8y\New-TestDeviceGroup
        $Group2 = PSc8y\New-TestDeviceGroup

    }

    It "Create group heirachy (parent group -> child group)" {
        $Response = PSc8y\New-ChildAssetReference -Group $Group1.id -NewChildGroup $Group2.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Group2.id
        PSc8y\Remove-ManagedObject -Id $Group1.id

    }
}

