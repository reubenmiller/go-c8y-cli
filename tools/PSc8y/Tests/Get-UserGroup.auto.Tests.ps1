. $PSScriptRoot/imports.ps1

Describe -Name "Get-UserGroup" {
    BeforeEach {
        $Group = New-TestUserGroup

    }

    It "Get a user group" {
        $Response = PSc8y\Get-UserGroup -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-UserGroup -Id $Group.id

    }
}

