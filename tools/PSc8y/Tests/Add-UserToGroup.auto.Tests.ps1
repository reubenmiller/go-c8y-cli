. $PSScriptRoot/imports.ps1

Describe -Name "Add-UserToGroup" {
    BeforeEach {
        $User = New-TestUser
        $Group = Get-UserGroupByName -Name "business"

    }

    It "Add a user to a user group" {
        $Response = PSc8y\Add-UserToGroup -Group $Group.id -User $User.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-User -Id $User.id

    }
}

