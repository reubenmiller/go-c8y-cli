. $PSScriptRoot/imports.ps1

Describe -Name "Add-RoleToUser" {
    BeforeEach {
        $User = PSc8y\New-TestUser -Name "customUser_"

    }

    It "Add a role (ROLE_ALARM_READ) to a user" {
        $Response = PSc8y\Add-RoleToUser -User $User.id -Role "ROLE_ALARM_READ"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Add a role to a user using wildcards" {
        $Response = PSc8y\Add-RoleToUser -User "customUser_*" -Role "*ALARM_*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Add a role to a user using wildcards (using pipeline)" {
        $Response = PSc8y\Get-RoleCollection -PageSize 100 | Where-Object Name -like "*ALARM*" | Add-RoleToUser -User "customUser_*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-User -Id $User.id

    }
}

