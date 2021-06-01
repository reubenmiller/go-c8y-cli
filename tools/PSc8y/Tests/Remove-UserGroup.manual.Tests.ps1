. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UserGroup" {
    Context "Existing groups" {
        BeforeAll {
            $Prefix = New-RandomString -Prefix "tempGroup1_"
            $Group1 = New-TestUserGroup -Name $Prefix
            $Group2 = New-TestUserGroup -Name $Prefix
        }

        It "Delete a multiple user groups (using pipeline)" {
            $Response = PSc8y\Get-UserGroupCollection -PageSize 2000 `
                | Where-Object { $_.name -like "${Prefix}*" } `
                | Remove-UserGroup

            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty

            # wait for deletion
            Start-Sleep -Seconds 5

            [array] $GroupsAfterDeletion = PSc8y\Get-UserGroupCollection -PageSize 2000 `
                | Where-Object { $_.name -like "${Prefix}*" }

            $GroupsAfterDeletion.Count | Should -BeExactly 0
        }

        AfterAll {
            $null = Remove-UserGroup -Id $Group1.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-UserGroup -Id $Group2.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
