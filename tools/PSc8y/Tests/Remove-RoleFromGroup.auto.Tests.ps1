. $PSScriptRoot/imports.ps1

Describe -Name "Remove-RoleFromGroup" {
    BeforeEach {
        $UserGroup = PSc8y\New-TestUserGroup
        Add-RoleToGroup -Group $UserGroup.id -Role "ROLE_MEASUREMENT_READ"

    }

    It "Remove a role from the given user group" {
        $Response = PSc8y\Remove-RoleFromGroup -Group $UserGroup.id -Role "ROLE_MEASUREMENT_READ"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-UserGroup -Id $UserGroup.id

    }
}

