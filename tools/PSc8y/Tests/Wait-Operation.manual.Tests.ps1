. $PSScriptRoot/imports.ps1

Describe -Name "Wait-Operation" {
    BeforeEach {
        $TestOperation = PSc8y\New-TestOperation
    }

    It "Wait for operation using invalid operaiton (fail fast)" {
        $StartTime = Get-Date
        $Response = "asdf8229d" | PSc8y\Wait-Operation -TimeoutSec 10 -WarningVariable "warning" -ErrorAction SilentlyContinue

        $warning | Should -Match "Could not find operation"
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeLessThan 5
    }

    It "Wait for operation (using pipeline)" {
        $Response = PSc8y\Get-Operation -Id $TestOperation.id | PSc8y\Wait-Operation -TimeoutSec 10 -WarningVariable "warnings"
        $warnings | Should -Match "Timeout: Operation is still being processed"
        $Response.id | Should -BeExactly $TestOperation.id
    }

    It "Wait for a successful operation (using pipeline)" {
        $TestOperation = $TestOperation.id | Update-Operation -Status SUCCESSFUL
        $Response = PSc8y\Get-Operation -Id $TestOperation.id | PSc8y\Wait-Operation -TimeoutSec 10 -WarningVariable "warnings"

        $warnings | Should -BeNullOrEmpty
        $Response.id | Should -BeExactly $TestOperation.id
        $Response.status | Should -BeExactly "SUCCESSFUL"
    }

    It "Wait for a failed operation (using pipeline)" {
        $TestOperation = $TestOperation.id | Update-Operation -Status FAILED
        $Response = PSc8y\Get-Operation -Id $TestOperation.id | PSc8y\Wait-Operation -TimeoutSec 10 -WarningVariable "warnings"

        $warnings | Should -Not -BeNullOrEmpty
        $Warnings | Should -Match "Reason"
        $Response.id | Should -BeExactly $TestOperation.id
        $Response.status | Should -BeExactly "FAILED"
    }

    AfterEach {
        if ($TestOperation.deviceId) {
            PSc8y\Remove-ManagedObject -Id $TestOperation.deviceId -ErrorAction SilentlyContinue
        }
    }
}
