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

    It "Creates a server user and subscribes to multiple tenants" {
        $Roles = @("ROLE_INVENTORY_READ", "ROLE_INVENTORY_ADMIN")
        $Response = PSc8y\New-ServiceUser -Name $AppName -Roles $Roles -Tenants t1111,t2222 -WhatIf 2>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response | Should -ContainRequest "POST /application/applications" -Total 1
        $Response | Should -ContainRequest "POST /tenant/tenants/t1111/applications" -Total 1
        $Response | Should -ContainRequest "POST /tenant/tenants/t2222/applications" -Total 1
        $Bodies = Get-RequestBodyCollection $Response
        $Bodies[0].name | Should -BeExactly $AppName
        $Bodies[0].requiredRoles | Should -BeExactly $Roles
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
