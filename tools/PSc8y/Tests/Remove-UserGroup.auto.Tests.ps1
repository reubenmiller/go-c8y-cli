. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UserGroup" {
    BeforeEach {
        $Group = New-TestUserGroup

    }

    It "Delete a user group" {
        $Response = PSc8y\Remove-UserGroup -Id $Group.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a user group (using pipeline)" {
        $Response = PSc8y\Get-UserGroupByName -Name $Group.name | Remove-UserGroup
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

