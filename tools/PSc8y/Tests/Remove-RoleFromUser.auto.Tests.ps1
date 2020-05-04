. $PSScriptRoot/imports.ps1

Describe -Name "Remove-RoleFromUser" {
    BeforeEach {
        $User = PSc8y\New-TestUser
        Add-RoleToUser -User $User.id -Role "ROLE_MEASUREMENT_READ"

    }

    It "Remove a role from the given user" {
        $Response = PSc8y\Remove-RoleFromUser -User $User.id -Role "ROLE_MEASUREMENT_READ"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        PSc8y\Remove-User -Id $User.id

    }
}

