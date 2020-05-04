. $PSScriptRoot/imports.ps1

Describe -Name "Register-Alias" {

    It "Aliases should be loaded by default when loading module" {
        $cmd = Get-Command "devices" -ErrorAction SilentlyContinue
        $cmd | Should -Not -BeNullOrEmpty
        $cmd.CommandType | Should -Be "Alias"
        Unregister-Alias
    }
}
