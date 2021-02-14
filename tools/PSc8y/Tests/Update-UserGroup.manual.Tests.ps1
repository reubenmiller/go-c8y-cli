. $PSScriptRoot/imports.ps1

Describe -Name "Update-UserGroup" {
    Context "Existing groups" {
        BeforeEach {
            $Group1 = New-TestUserGroup -Name "tempGroup1"
        }

        It "Get a group (using pipeline)" {
            $NewName = New-RandomString -Prefix "updateGroupName1"
            $Response = $Group1 | PSc8y\Update-UserGroup -Name $NewName

            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.name | Should -BeExactly $NewName
        }

        AfterEach {
            $null = Remove-UserGroup -Id $Group1.id
        }
    }
}
