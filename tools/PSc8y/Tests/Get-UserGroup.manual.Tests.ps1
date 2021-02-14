. $PSScriptRoot/imports.ps1

Describe -Name "Get-UserGroup" {
    Context "Existing groups" {
        BeforeEach {
            $Group1 = New-TestUserGroup -Name "tempGroup1"
            $Group2 = New-TestUserGroup -Name "tempGroup1"
        }

        It "Get a group (using pipeline)" {
            $Response = $Group1, $Group2 | PSc8y\Get-UserGroup

            $LASTEXITCODE | Should -Be 0
            $Response | Should -HaveCount 2
            $Response[0].id | Should -BeExactly $Group1.id
            $Response[1].id | Should -BeExactly $Group2.id
        }

        AfterEach {
            $null = Remove-UserGroup -Id $Group1.id
            $null = Remove-UserGroup -Id $Group2.id
        }
    }
}
