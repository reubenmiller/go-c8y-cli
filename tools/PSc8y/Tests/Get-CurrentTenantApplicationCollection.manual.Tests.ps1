
. $PSScriptRoot/imports.ps1

Describe -Name "Get-CurrentTenantApplicationCollection" {
    It "Get a list of applications in current tenant" {
        [array] $Response = PSc8y\Get-CurrentTenantApplicationCollection
        $Response | Should -Not -BeNullOrEmpty
        $Response.Count | Should -BeGreaterThan 0
        $Response[0].id | Should -Not -BeNullOrEmpty
    }
}
