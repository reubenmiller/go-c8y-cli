. $PSScriptRoot/imports.ps1

Describe -Name "Get-BulkOperation" {
    BeforeEach {
        $Group = New-TestDeviceGroup -TotalDevices 2
        $BulkOp = New-BulkOperation -Group $Group.id -CreationRampSec 10 -Operation @{c8y_Restart=@{}}

    }

    It "Get bulk operation by id" {
        $Response = PSc8y\Get-BulkOperation -Id $BulkOp.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-BulkOperationCollection | Remove-BulkOperation
        Remove-DeviceGroup -Id $Group.id -Cascade

    }
}

