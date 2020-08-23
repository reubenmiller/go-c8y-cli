. $PSScriptRoot/imports.ps1

Describe -Name "Get-BulkOperation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation

    }

    It "Get bulk operation by id" {
        $Response = PSc8y\Get-BulkOperation -Id $TestOperation.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }

    }
}

