. $PSScriptRoot/imports.ps1

Describe -Name "Get-Operation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation

    }

    It "Get operation by id" {
        $Response = PSc8y\Get-Operation -Id $TestOperation.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }

    }
}

