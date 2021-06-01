. $PSScriptRoot/imports.ps1

Describe -Name "Add-UserToGroup" {
    BeforeEach {
        $User = New-TestUser
        $Group = Get-UserGroupByName -Name "devices"

    }

    It "Add a user to a user group using pipeline" {
        $Response = $User | PSc8y\Add-UserToGroup -Group $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-User -Id $User.id

    }
}
