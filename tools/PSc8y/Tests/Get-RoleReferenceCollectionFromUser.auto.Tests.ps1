. $PSScriptRoot/imports.ps1

Describe -Name "Get-RoleReferenceCollectionFromUser" {
    BeforeEach {
        $User = New-TestUser
        Add-RoleToUser -User $User.id -Role "ROLE_ALARM_READ"

    }

    It "Get a list of role references for a user" {
        $Response = PSc8y\Get-RoleReferenceCollectionFromUser -User $User.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-User -Id $User.id

    }
}

