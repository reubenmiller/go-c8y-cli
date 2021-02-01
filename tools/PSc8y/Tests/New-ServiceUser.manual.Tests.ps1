. $PSScriptRoot/imports.ps1

Describe -Name "New-ServiceUser" -Tag "microservice" {
    BeforeEach {
        $AppName = (New-RandomString -Prefix "testms-")
    }

    It "Creates a new microservice service user application" {
        $Roles = @("ROLE_INVENTORY_READ")
        $Response = PSc8y\New-ServiceUser -Name $AppName -Roles $Roles
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.name | Should -BeExactly $AppName
        $Response.requiredRoles | Should -BeExactly $Roles
    }

    It "Creates a new microservice service user application with more than 1 role" {
        $Roles = @("ROLE_INVENTORY_READ", "ROLE_INVENTORY_ADMIN")
        $Response = PSc8y\New-ServiceUser -Name $AppName -Roles $Roles
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.name | Should -BeExactly $AppName
        $Response.requiredRoles | Should -BeExactly $Roles
    }

    AfterEach {
        Remove-Microservice -Id $AppName
    }
}
