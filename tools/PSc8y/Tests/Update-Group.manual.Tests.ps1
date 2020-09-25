. $PSScriptRoot/imports.ps1

Describe -Name "Update-Group" {
    Context "Existing groups" {
        BeforeEach {
            $Group1 = New-TestGroup -Name "tempGroup1"
        }

        It "Get a group (using pipeline)" {
            $NewName = New-RandomString -Prefix "updateGroupName1"
            $Response = $Group1 | PSc8y\Update-Group -Name $NewName

            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.name | Should -BeExactly $NewName
        }

        AfterEach {
            $null = Remove-Group -Id $Group1.id
        }
    }
}
