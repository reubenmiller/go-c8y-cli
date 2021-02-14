. $PSScriptRoot/imports.ps1

Describe -Name "Get-UserGroupByName" {
    BeforeEach {
        $Group = New-TestUserGroup

    }

    It "Get user group by its name" {
        $Response = PSc8y\Get-UserGroupByName -Name $Group.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-UserGroup -Id $Group.id

    }
}

