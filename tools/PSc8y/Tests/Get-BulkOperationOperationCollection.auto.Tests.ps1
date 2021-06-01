. $PSScriptRoot/imports.ps1

Describe -Name "Get-BulkOperationOperationCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of pending operations from bulk operation with id 10" {
        $Response = PSc8y\Get-BulkOperationOperationCollection -Id 10 -Status PENDING
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Get all pending operations from all bulk operations which are still in progress (using pipeline)" {
        $Response = PSc8y\Get-BulkOperationCollection | Where-Object {     It -Skip "Get all pending operations from all bulk operations which are still in progress (using pipeline)" {
        $Response = PSc8y\{{ Command }}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }
.status -eq "IN_PROGRESS" } | Get-BulkOperationOperationCollection -Status PENDING
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It -Skip "Check all bulk operations if they have any related operations still in executing state and were created more than 10 days ago, then cancel it with a custom message" {
        $Response = PSc8y\Get-BulkOperationCollection | Get-BulkOperationOperationCollection -status EXECUTING --dateTo "-10d" | Update-Operation -Status FAILED -FailureReason "Manually cancelled stale operation"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

