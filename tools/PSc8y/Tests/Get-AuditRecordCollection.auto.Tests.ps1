. $PSScriptRoot/imports.ps1

Describe -Name "Get-AuditRecordCollection" {
    BeforeEach {
        $Device1 = New-TestDevice
        $Device2 = New-TestDevice
        Remove-ManagedObject -Id $Device2.id
        $Device3 = New-TestAgent
        $Operation = New-TestOperation -Device $Device3
        Update-Operation -Id $Operation.id -Status "EXECUTING"
        Update-Operation -Id $Operation.id -Status "SUCCESSFUL"

    }

    It "Get a list of audit records" {
        $Response = PSc8y\Get-AuditRecordCollection -PageSize 100
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of audit records related to a managed object" {
        $Response = PSc8y\Get-AuditRecordCollection -Source $Device2.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of audit records related to an operation" {
        $Response = PSc8y\Get-Operation -Id $Operation.id | Get-AuditRecordCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device1.id
        Remove-ManagedObject -Id $Device3.id

    }
}

