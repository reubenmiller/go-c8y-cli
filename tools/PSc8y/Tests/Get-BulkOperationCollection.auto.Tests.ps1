. $PSScriptRoot/imports.ps1

Describe -Name "Get-BulkOperationCollection" {
    BeforeEach {
        $Group = New-TestDeviceGroup -TotalDevices 2
        $BulkOp = New-BulkOperation -Group $Group.id -CreationRampSec 10 -Operation @{c8y_Restart=@{}}

    }

    It -Skip "Get a list of bulk operations" {
        $Response = PSc8y\Get-BulkOperationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get a list of bulk operations created in the last 1 day" {
        $Response = PSc8y\Get-BulkOperationCollection -DateFrom -1d
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get a list of bulk operations in the general status SCHEDULED or EXECUTING" {
        $Response = PSc8y\Get-BulkOperationCollection -Status SCHEDULED, EXECUTING
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-BulkOperationCollection | Remove-BulkOperation
        Remove-DeviceGroup -Id $Group.id -Cascade

    }
}

