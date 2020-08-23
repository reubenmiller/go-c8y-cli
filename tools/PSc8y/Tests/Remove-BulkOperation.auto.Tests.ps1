. $PSScriptRoot/imports.ps1

Describe -Name "Remove-BulkOperation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation

    }

    It "Remove bulk operation by id" {
        $Response = PSc8y\Remove-BulkOperation -Id $TestOperation.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }

    }
}

