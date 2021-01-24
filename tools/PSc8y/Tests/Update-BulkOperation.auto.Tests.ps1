. $PSScriptRoot/imports.ps1

Describe -Name "Update-BulkOperation" {
    BeforeEach {
        $Group = New-TestDeviceGroup -TotalDevices 2
        $BulkOp = New-BulkOperation -Group $Group.id -CreationRampSec 10 -Operation @{c8y_Restart=@{}}

    }

    It "Update bulk operation wait period between the creation of each operation to 1.5 seconds" {
        $Response = PSc8y\Update-BulkOperation -Id $BulkOp.id -CreationRamp 1.5
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-BulkOperationCollection | Remove-BulkOperation
        Remove-DeviceGroup -Id $Group.id -Cascade

    }
}

