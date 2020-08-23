. $PSScriptRoot/imports.ps1

Describe -Name "Update-BulkOperation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation

    }

    It "Update bulk operation" {
        $Response = PSc8y\Update-BulkOperation -Id $TestOperation.id -CreationRamp 15
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }

    }
}

