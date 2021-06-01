. $PSScriptRoot/imports.ps1

Describe -Name "Wait-Operation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation
    }

    It "Wait for operation using invalid operation (fail fast)" {
        $StartTime = Get-Date
        $warning = $( $Response = "asdf8229d" | PSc8y\Wait-Operation -Duration "10s" ) 2>&1
        $LASTEXITCODE | Should -BeExactly 100
        $Response | Should -BeNullOrEmpty
        ($warning -join "`n") | Should -Match "No operation for"
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeLessThan 5
    }

    It "Wait for operation (using pipeline)" {
        $warnings = $( $Response = PSc8y\Get-Operation -Id $TestOperation.id | PSc8y\Wait-Operation -Duration "10s" ) 2>&1
        $warnings | Should -Match "Timeout"
        $Response.id | Should -BeExactly $TestOperation.id
    }

    It "Wait for a successful operation (using pipeline)" {
        $TestOperation = $TestOperation.id | Update-Operation -Status SUCCESSFUL
        $Response = PSc8y\Get-Operation -Id $TestOperation.id | PSc8y\Wait-Operation -Duration "10s" -WarningVariable "warnings"

        $warnings | Should -BeNullOrEmpty
        $Response.id | Should -BeExactly $TestOperation.id
        $Response.status | Should -BeExactly "SUCCESSFUL"
    }

    It "Wait for a failed operation (using pipeline)" {
        $TestOperation = $TestOperation.id | Update-Operation -Status FAILED -FailureReason "some error"
        $warnings = $( $Response = PSc8y\Get-Operation -Id $TestOperation.id | PSc8y\Wait-Operation -Duration "10s" ) 2>&1

        $warnings | Should -Not -BeNullOrEmpty
        $Warnings[0] | Should -Match "Reason"
        $Response.id | Should -BeExactly $TestOperation.id
        $Response.status | Should -BeExactly "FAILED"
    }

    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }
    }
}
