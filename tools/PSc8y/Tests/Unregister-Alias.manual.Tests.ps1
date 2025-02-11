. $PSScriptRoot/imports.ps1

Describe -Name "Unregister-Alias" {

    It "Removes aliases" {
        Unregister-Alias
        $cmd = Get-Command "devices" -ErrorAction SilentlyContinue
        $cmd | Should -BeNullOrEmpty
    }
}
