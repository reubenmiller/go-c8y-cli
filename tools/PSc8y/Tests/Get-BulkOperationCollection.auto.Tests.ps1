. $PSScriptRoot/imports.ps1

Describe -Name "Get-BulkOperationCollection" {
    BeforeEach {
        $Group = New-TestDeviceGroup -TotalDevices 2
        $BulkOp = New-BulkOperation -Group $Group.id -CreationRampSec 10 -Operation @{c8y_Restart=@{}}

    }

    It "Get a list of bulk operations" {
        $Response = PSc8y\Get-BulkOperationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-BulkOperationCollection | Remove-BulkOperation
        Remove-DeviceGroup -Id $Group.id -Cascade

    }
}

