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
        $output = PSc8y\New-ServiceUser -Name $AppName -Roles $Roles -Tenants t1111,t2222 -Dry -DryFormat json 2>&1
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $Requests = $output | ConvertFrom-Json

        $Requests | Should -HaveCount 3
        $Requests[0] | Should -MatchObject @{method="POST"; path="/application/applications"} -Property method, path
        $Requests[1] | Should -MatchObject @{method="POST"; path="/tenant/tenants/t1111/applications"} -Property method, path
        $Requests[2] | Should -MatchObject @{method="POST"; path="/tenant/tenants/t2222/applications"} -Property method, path

        $Requests[0].body.name | Should -BeExactly $AppName
        $Requests[0].body.requiredRoles | Should -BeExactly $Roles
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
