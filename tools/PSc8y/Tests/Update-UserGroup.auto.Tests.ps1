. $PSScriptRoot/imports.ps1

Describe -Name "Update-UserGroup" {
    BeforeEach {
        $Group = New-TestUserGroup

    }

    It "Update a user group" {
        $Response = PSc8y\Update-UserGroup -Id $Group -Name "customGroup2"
        $LASTEXITCODE | Should -Be 0
    }

    It "Update a user group (using pipeline)" {
        $Response = PSc8y\Get-UserGroupByName -Name $Group.name | Update-UserGroup -Name "customGroup2"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-UserGroup -Id $Group.id

    }
}

