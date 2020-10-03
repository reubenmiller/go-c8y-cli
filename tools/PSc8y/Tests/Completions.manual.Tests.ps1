. $PSScriptRoot/imports.ps1

Describe -Name "Create-Completions" {

    BeforeEach {
        $originalSetting = $env:C8Y_SESSION
        $env:C8Y_SESSION = ""
    }

    It "Create bash completions" {
        $c8y = PSc8y\Get-ClientBinary
        $output = & $c8y completion bash
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
    }

    It "Create zsh completions" {
        $c8y = PSc8y\Get-ClientBinary
        $output = & $c8y completion zsh
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
    }

    It "Create powershell completions" {
        $c8y = PSc8y\Get-ClientBinary
        $output = & $c8y completion powershell
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
    }

    AfterEach {
        $env:C8Y_SESSION = $originalSetting
    }
}
