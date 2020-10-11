. $PSScriptRoot/imports.ps1

Describe -Name "Get-ClientSetting" {

    It "shows a list of client settings" {
        $Response = Get-ClientSetting
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }
}
