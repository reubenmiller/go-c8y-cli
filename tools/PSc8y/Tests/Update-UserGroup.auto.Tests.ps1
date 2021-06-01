. $PSScriptRoot/imports.ps1

Describe -Name "Update-UserGroup" {
    BeforeEach {
        $Group = New-TestUserGroup
        $GroupName = New-RandomString -Prefix "customGroup2"
        $NewGroupName = New-RandomString -Prefix "newCustomGroup2"

    }

    It "Update a user group" {
        $Response = PSc8y\Update-UserGroup -Id $Group -Name $GroupName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update a user group (using pipeline)" {
        $Response = PSc8y\Get-UserGroupByName -Name $Group.name | Update-UserGroup -Name $NewGroupName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-UserGroup -Id $Group.id

    }
}

