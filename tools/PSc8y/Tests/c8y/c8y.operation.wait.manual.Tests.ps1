. $PSScriptRoot/../imports.ps1

Describe -Name "Wait-Operation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation
    }

    It "Wait for operation using invalid operation (fail fast)" {
        $StartTime = Get-Date
        $Response = "asdf8229d" | c8y operations wait --duration "10s"
        $LASTEXITCODE | Should -BeExactly 100
        $Response | Should -BeNullOrEmpty
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeLessThan 5
    }

    It "Wait for operation (using pipeline)" {
        $output = c8y operations get --id $TestOperation.id | c8y operations wait --duration "10s"
        $LASTEXITCODE | Should -BeExactly 106
        $Response = $output | ConvertFrom-Json
        # $warnings | Should -Match "Timeout: Operation is still being processed"
        $Response.id | Should -BeExactly $TestOperation.id
    }

    It "Wait for a successful operation (using pipeline)" {
        $TestOperation = $TestOperation.id | c8y operations update --status SUCCESSFUL | ConvertFrom-Json
        $output = c8y operations get --id $TestOperation.id | c8y operations wait --duration "10s"
        $response = $output | ConvertFrom-Json
        $warnings | Should -BeNullOrEmpty
        $response.id | Should -BeExactly $TestOperation.id
        $response.status | Should -BeExactly "SUCCESSFUL"
    }

    It "Wait for a failed operation (using pipeline)" {
        $TestOperation = $TestOperation.id | c8y operations update --status FAILED --failureReason "my root cause" | ConvertFrom-Json
        $warnings = $( $output = c8y operations get --id $TestOperation.id | c8y operations wait --duration "10s" ) 2>&1
        $response = $output | ConvertFrom-Json

        $warnings | Should -Not -BeNullOrEmpty
        $Warnings[0] | Should -Match "failure reason: my root cause"
        $response.id | Should -BeExactly $TestOperation.id
        $response.status | Should -BeExactly "FAILED"
    }

    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }
    }
}
