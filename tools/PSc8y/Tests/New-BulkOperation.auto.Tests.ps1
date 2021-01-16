. $PSScriptRoot/imports.ps1

Describe -Name "New-BulkOperation" {
    BeforeEach {
        $Group = New-TestDeviceGroup -TotalDevices 2

    }

    It "Create bulk operation for a group" {
        $Response = PSc8y\New-BulkOperation -Group $Group.id -StartDate "60s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create bulk operation for a group (using pipeline)" {
        $Response = PSc8y\Get-DeviceGroup $Group.id | New-BulkOperation -StartDate "10s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-BulkOperationCollection | Remove-BulkOperation
        Remove-DeviceGroup -Id $Group.id -Cascade

    }
}

