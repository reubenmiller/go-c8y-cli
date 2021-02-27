. $PSScriptRoot/imports.ps1

Describe -Name "Get-ServiceUser" -Tag "microservice" {
    BeforeEach {
        $AppName = (New-RandomString -Prefix "testms-")
        $CurrentTenant = Get-CurrentTenant
        $App = New-ServiceUser -Name $AppName -Tenants $CurrentTenant.name
    }

    It "Get microservice service user" {
        $Response = PSc8y\Get-ServiceUser -Id $AppName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.name | Should -Not -BeNullOrEmpty
        $Response.password | Should -Not -BeNullOrEmpty
        $Response.tenant | Should -Not -BeNullOrEmpty
    }

    It "Get microservice service user using pipeline" {
        $Response = Get-Application -Id $App.id | PSc8y\Get-ServiceUser
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.name | Should -Not -BeNullOrEmpty
        $Response.password | Should -Not -BeNullOrEmpty
        $Response.tenant | Should -Not -BeNullOrEmpty
    }

    AfterEach {
        if ($App.id) {
            Remove-Microservice -Id $App.id
        }
    }
}
