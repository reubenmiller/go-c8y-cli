. $PSScriptRoot/../imports.ps1

Describe -Name "c8y session" {
    It "Get session home folder" {
        $output = c8y settings list --select session.home --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -BeExactly $env:C8Y_SESSION_HOME
    }
}
