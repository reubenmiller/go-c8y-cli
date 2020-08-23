. $PSScriptRoot/imports.ps1

Describe -Name "New-BulkOperation" {
    BeforeEach {
        $group = New-TestAgent

    }

    It "Create bulk operation for a group" {
        $Response = PSc8y\New-BulkOperation -Group $group.id -StartDate "10s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create bulk operation for a group (using pipeline)" {
        $Response = PSc8y\Get-DeviceGroup $group.id | New-BulkOperation -StartDate "10s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $group.id

    }
}

