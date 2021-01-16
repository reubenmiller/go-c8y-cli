. $PSScriptRoot/imports.ps1

Describe -Name "Remove-BulkOperation" {
    BeforeEach {
        $Group = New-TestDeviceGroup -TotalDevices 2
        $BulkOp = New-BulkOperation -Group $Group.id -CreationRampSec 10 -Operation @{c8y_Restart=@{}}

    }

    It "Remove bulk operation by id" {
        $Response = PSc8y\Remove-BulkOperation -Id $BulkOp.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-DeviceGroup -Id $Group.id -Cascade

    }
}

