. $PSScriptRoot/imports.ps1

Describe -Name "Invoke-CancelOperation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation
        $Agent = PSc8y\New-TestAgent
        $Operation1 = PSc8y\New-TestOperation -Device $Agent.id
        $Operation2 = PSc8y\New-TestOperation -Device $Agent.id

    }

    It "Cancel an operation" {
        $Response = PSc8y\Invoke-CancelOperation -Id $TestOperation.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Cancel multiple operations" {
        $Response = PSc8y\Get-OperationCollection -Device $Agent.id -Status PENDING | Invoke-CancelOperation -FailureReason "manually cancelled"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }
        PSc8y\Remove-ManagedObject -Id $Agent.id

    }
}

