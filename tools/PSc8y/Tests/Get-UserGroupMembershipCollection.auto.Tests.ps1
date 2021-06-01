. $PSScriptRoot/imports.ps1

Describe -Name "Get-UserGroupMembershipCollection" {
    BeforeEach {
        $User = New-TestUser
        $Group = Get-UserGroupByName -Name "business"
        Add-UserToGroup -Group $Group.id -User $User.id

    }

    It "List the users within a user group" {
        $Response = PSc8y\Get-UserGroupMembershipCollection -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "List the users within a user group (using pipeline)" {
        $Response = PSc8y\Get-UserGroupByName -Name "business" | Get-UserGroupMembershipCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-User -Id $User.id

    }
}

